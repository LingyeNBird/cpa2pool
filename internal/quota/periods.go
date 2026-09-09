package quota

import (
 "database/sql"
 "errors"
 "time"
 "cpa2pool/internal/domain"
 "cpa2pool/internal/store"
)

const stamp="2006-01-02T15:04:05.000000000Z"
func text(t time.Time)string{return t.UTC().Format(stamp)}
func nullable(t *time.Time)any{if t==nil{return nil};return text(*t)}
func parse(s sql.NullString)*time.Time{if !s.Valid{return nil};t,_:=time.Parse(stamp,s.String);return &t}
func loadPeriod(q store.Query,query string,args ...any)(domain.Period,error){var p domain.Period;var start string;var end,closed sql.NullString;err:=q.QueryRow(query,args...).Scan(&p.ID,&p.QuotaID,&p.ParticipantID,&start,&end,&closed,&p.Limit,&p.Used);p.StartsAt,_=time.Parse(stamp,start);p.EndsAt=parse(end);p.ClosedAt=parse(closed);p.Remaining=p.Limit-p.Used;return p,err}
func Get(q store.Query,id string)(domain.Quota,error){return store.One[domain.Quota](q,"SELECT body FROM quotas WHERE id=?",id)}
func List(q store.Query,pid string)([]domain.Quota,error){return store.List[domain.Quota](q,"SELECT body FROM quotas WHERE participant_id=? ORDER BY rowid",pid)}

// Monthly boundaries retain the original day, clamped only in shorter months.
func nextBoundary(q domain.Quota,start time.Time)*time.Time{
 var end time.Time
 switch q.Period{case "none":return nil;case "day":end=start.AddDate(0,0,1);case "week":end=start.AddDate(0,0,7);case "month":y,m,_:=start.Date();m++;last:=time.Date(y,m+1,0,0,0,0,0,time.UTC).Day();day:=q.Anchor.Day();if day>last{day=last};end=time.Date(y,m,day,q.Anchor.Hour(),q.Anchor.Minute(),q.Anchor.Second(),q.Anchor.Nanosecond(),time.UTC)}
 return &end
}
func createPeriod(tx *sql.Tx,q domain.Quota,start time.Time)(domain.Period,error){p:=domain.Period{ID:domain.ID(),QuotaID:q.ID,ParticipantID:q.ParticipantID,StartsAt:start,EndsAt:nextBoundary(q,start),Limit:q.Limit,Remaining:q.Limit};if q.ExpiresAt!=nil&&(p.EndsAt==nil||q.ExpiresAt.Before(*p.EndsAt)){p.EndsAt=q.ExpiresAt};_,err:=tx.Exec("INSERT INTO periods VALUES(?,?,?,?,?,?,?,?)",p.ID,q.ID,q.ParticipantID,text(start),nullable(p.EndsAt),nil,int64(p.Limit),0);return p,err}
func Current(tx *sql.Tx,q domain.Quota,now time.Time)(domain.Period,error){
 p,err:=loadPeriod(tx,"SELECT * FROM periods WHERE quota_id=? AND closed_at IS NULL",q.ID)
 if err==sql.ErrNoRows{
  if q.ExpiresAt!=nil&&!now.Before(*q.ExpiresAt){
   p,err=loadPeriod(tx,"SELECT * FROM periods WHERE quota_id=? ORDER BY rowid DESC LIMIT 1",q.ID)
   if err==nil{return p,nil}
  }
  p,err=createPeriod(tx,q,q.StartsAt)
 };if err!=nil{return p,err}
 for p.EndsAt!=nil&&!now.Before(*p.EndsAt){
  boundary:=*p.EndsAt
  if _,err=tx.Exec("UPDATE periods SET closed_at=? WHERE id=?",text(boundary),p.ID);err!=nil{return p,err}
  if err=store.Audit(tx,q.ParticipantID,q.ID,"period.end","",p,nil);err!=nil{return p,err}
  if q.Period=="none"||q.ExpiresAt!=nil&&!boundary.Before(*q.ExpiresAt){p.ClosedAt=&boundary;return p,nil}
  p,err=createPeriod(tx,q,boundary);if err!=nil{return p,err}
 }
 return p,nil
}
func Active(q domain.Quota,now time.Time)bool{return q.Enabled&&!now.Before(q.StartsAt)&&(q.ExpiresAt==nil||now.Before(*q.ExpiresAt))}
func History(q store.Query,pid,qid string)([]domain.Period,error){
 rows,err:=q.Query("SELECT id FROM periods WHERE (?='' OR participant_id=?) AND (?='' OR quota_id=?) ORDER BY starts_at DESC",pid,pid,qid,qid);if err!=nil{return nil,err};ids:=[]string{};for rows.Next(){var id string;if err=rows.Scan(&id);err!=nil{rows.Close();return nil,err};ids=append(ids,id)};err=rows.Err();rows.Close();if err!=nil{return nil,err};out:=[]domain.Period{};for _,id:=range ids{p,e:=loadPeriod(q,"SELECT * FROM periods WHERE id=?",id);if e!=nil{return nil,e};out=append(out,p)};return out,nil
}
func valid(q domain.Quota)error{if q.Name==""||q.ParticipantID==""{return errors.New("额度名称与参与者不能为空")};if q.Limit<0{return errors.New("限额不能为负")};switch q.Period{case "none","day","week","month":default:return errors.New("无效额度周期")};if q.ExpiresAt!=nil&&!q.ExpiresAt.After(q.StartsAt){return errors.New("有效期必须晚于开始时间")};return nil}

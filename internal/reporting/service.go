package reporting

import (
 "fmt"
 "strings"
 "cpa2pool/internal/domain"
 "cpa2pool/internal/quota"
 "cpa2pool/internal/store"
)
type Service struct{ Store *store.Store }
type Filter struct { ParticipantID string; Model string; From string; To string; Limit int; Offset int }
func(f Filter) where()(string,[]any){parts:=[]string{"1=1"};args:=[]any{};for _,x:=range []struct{column,value,op string}{{"participant_id",f.ParticipantID,"="},{"model",f.Model,"="},{"time",f.From,">="},{"time",f.To,"<"}}{if x.value!=""{parts=append(parts,x.column+x.op+"?");args=append(args,x.value)}};return strings.Join(parts," AND "),args}
type Page[T any] struct{Items []T `json:"items"`; Total int64 `json:"total"`}
func(s Service) Bills(f Filter)(Page[domain.Bill],error){var p Page[domain.Bill];w,args:=f.where();if err:=s.Store.DB.QueryRow("SELECT COUNT(*) FROM bills WHERE "+w,args...).Scan(&p.Total);err!=nil{return p,err};items,e:=store.List[domain.Bill](s.Store.DB,"SELECT body FROM bills WHERE "+w+" ORDER BY time DESC LIMIT ? OFFSET ?",append(args,f.Limit,f.Offset)...);p.Items=items;return p,e}
type Stat struct{ Group string `json:"group"`; Requests int64 `json:"requests"`; Cost domain.Money `json:"cost"`; Input int64 `json:"input"`; Output int64 `json:"output"`; CacheRead int64 `json:"cache_read"` }
func(s Service) Stats(f Filter,group string)([]Stat,error){column:="'all'";switch group{case "participant":column="participant_id";case "model":column="model";case "day":column="substr(time,1,10)";case "","all":default:return nil,fmt.Errorf("无效统计维度")};w,args:=f.where();rows,e:=s.Store.DB.Query("SELECT "+column+",COUNT(*),COALESCE(SUM(cost),0),COALESCE(SUM(json_extract(body,'$.usage.input')),0),COALESCE(SUM(json_extract(body,'$.usage.output')),0),COALESCE(SUM(json_extract(body,'$.usage.cache_read')),0) FROM bills WHERE "+w+" GROUP BY "+column+" ORDER BY 1",args...);if e!=nil{return nil,e};defer rows.Close();out:=[]Stat{};for rows.Next(){var v Stat;if e=rows.Scan(&v.Group,&v.Requests,&v.Cost,&v.Input,&v.Output,&v.CacheRead);e!=nil{return nil,e};out=append(out,v)};return out,rows.Err()}
func(s Service) Audits(f Filter)(Page[domain.Audit],error){var p Page[domain.Audit];f.Model="";w,args:=f.where();if err:=s.Store.DB.QueryRow("SELECT COUNT(*) FROM audits WHERE "+w,args...).Scan(&p.Total);err!=nil{return p,err};items,e:=store.List[domain.Audit](s.Store.DB,"SELECT body FROM audits WHERE "+w+" ORDER BY time DESC LIMIT ? OFFSET ?",append(args,f.Limit,f.Offset)...);p.Items=items;return p,e}
func(s Service) Periods(pid,qid string)([]domain.Period,error){return quota.History(s.Store.DB,pid,qid)}

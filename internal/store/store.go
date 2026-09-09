package store

import (
 "database/sql"
 "encoding/json"
 "os"
 "path/filepath"
 "cpa2pool/internal/domain"
 _ "modernc.org/sqlite"
)

type Store struct { DB *sql.DB }
type Query interface { Exec(string,...any)(sql.Result,error); Query(string,...any)(*sql.Rows,error); QueryRow(string,...any)*sql.Row }
func Open(path string)(*Store,error){
 if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{return nil,err}
 db,err:=sql.Open("sqlite",path);if err!=nil{return nil,err};db.SetMaxOpenConns(1)
 _,err=db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;
 CREATE TABLE IF NOT EXISTS participants(id TEXT PRIMARY KEY,scope TEXT UNIQUE NOT NULL,body TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS prices(model TEXT PRIMARY KEY,body TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS quotas(id TEXT PRIMARY KEY,participant_id TEXT NOT NULL REFERENCES participants(id),body TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS periods(id TEXT PRIMARY KEY,quota_id TEXT NOT NULL REFERENCES quotas(id),participant_id TEXT NOT NULL,starts_at TEXT NOT NULL,ends_at TEXT,closed_at TEXT,limit_amount INTEGER NOT NULL,used INTEGER NOT NULL DEFAULT 0);
 CREATE UNIQUE INDEX IF NOT EXISTS one_open_period ON periods(quota_id) WHERE closed_at IS NULL;
 CREATE TABLE IF NOT EXISTS bills(id TEXT PRIMARY KEY,request_id TEXT UNIQUE NOT NULL,participant_id TEXT NOT NULL,model TEXT NOT NULL,time TEXT NOT NULL,cost INTEGER NOT NULL,body TEXT NOT NULL);
 CREATE INDEX IF NOT EXISTS bill_filter ON bills(participant_id,time,model);
 CREATE TABLE IF NOT EXISTS audits(id TEXT PRIMARY KEY,participant_id TEXT NOT NULL,quota_id TEXT NOT NULL,time TEXT NOT NULL,body TEXT NOT NULL);
 CREATE INDEX IF NOT EXISTS audit_filter ON audits(participant_id,time);
 PRAGMA user_version=1;`)
 if err!=nil{db.Close();return nil,err};return &Store{DB:db},nil
}
func(s *Store) Tx(fn func(*sql.Tx)error)error { tx,err:=s.DB.Begin();if err!=nil{return err};defer tx.Rollback();if err=fn(tx);err!=nil{return err};return tx.Commit() }
func JSON(v any)string { b,err:=json.Marshal(v);if err!=nil{panic(err)};return string(b) }
func One[T any](q Query,query string,args ...any)(T,error){var v T;var b string;err:=q.QueryRow(query,args...).Scan(&b);if err!=nil{return v,err};err=json.Unmarshal([]byte(b),&v);return v,err}
func List[T any](q Query,query string,args ...any)([]T,error){rows,err:=q.Query(query,args...);if err!=nil{return nil,err};defer rows.Close();out:=[]T{};for rows.Next(){var b string;var v T;if err=rows.Scan(&b);err!=nil{return nil,err};if err=json.Unmarshal([]byte(b),&v);err!=nil{return nil,err};out=append(out,v)};return out,rows.Err()}
func Audit(q Query,pid,qid,action,note string,before,after any)error { a:=domain.Audit{ID:domain.ID(),ParticipantID:pid,QuotaID:qid,Action:action,Note:note,Time:domain.Now(),Before:before,After:after};_,err:=q.Exec("INSERT INTO audits VALUES(?,?,?,?,?)",a.ID,pid,qid,a.Time.Format("2006-01-02T15:04:05.000000000Z"),JSON(a));return err }

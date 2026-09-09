package plugin

import (
 "encoding/json"
 "errors"
 "net/http"
 "sync"
 "cpa2pool/internal/api"
 "cpa2pool/internal/billing"
 "cpa2pool/internal/console"
 "cpa2pool/internal/store"
 "gopkg.in/yaml.v3"
)

type Runtime struct { mu sync.Mutex; db *store.Store; billing *billing.Service; path string }
type execution struct { RequestID string; Model string; RequestedModel string; Body []byte; Payload []byte; Metadata map[string]any; EventType string; ChunkIndex int }
func(r *Runtime) Close(){r.mu.Lock();defer r.mu.Unlock();if r.db!=nil{r.db.DB.Close();r.db=nil}}
func(r *Runtime) Register(raw []byte)(any,error){r.mu.Lock();defer r.mu.Unlock();var req struct{ConfigYAML []byte `json:"config_yaml"`};if err:=json.Unmarshal(raw,&req);err!=nil{return nil,err};var cfg struct{Database string `yaml:"database"`};if err:=yaml.Unmarshal(req.ConfigYAML,&cfg);err!=nil{return nil,err};if cfg.Database==""{cfg.Database="data/cpa2pool.db"};if r.db!=nil&&r.path!=cfg.Database{return nil,errors.New("修改数据库路径需要重启 CPA")};if r.db==nil{db,err:=store.Open(cfg.Database);if err!=nil{return nil,err};r.db=db;r.path=cfg.Database;r.billing=billing.New(db)}
 return map[string]any{"schema_version":6,"metadata":map[string]any{"Name":"拼车额度","Version":"0.1.0","Author":"cpa2pool","GitHubRepository":"local:cpa2pool","ConfigFields":[]any{}},"capabilities":map[string]bool{"management_api":true,"request_interceptor":true,"request_lifecycle_plugin":true,"response_interceptor":true,"response_stream_interceptor":true,"websocket_response_observer":true}},nil
}
func(r *Runtime) Handle(method string,raw []byte)(any,error){
 switch method{
 case "plugin.register","plugin.reconfigure":return r.Register(raw)
 case "plugin.quiesce":return map[string]any{},nil
 case "plugin.shutdown":r.Close();return map[string]any{},nil
 case "management.register":return map[string]any{"routes":api.Routes(),"resources":[]map[string]string{{"Path":"/ui","Menu":"拼车额度"}}},nil
 case "management.handle":var req api.Request;if err:=json.Unmarshal(raw,&req);err!=nil{return nil,err};if req.Path=="/v0/resource/plugins/cpa2pool/ui"{return api.Response{StatusCode:200,Headers:http.Header{"Content-Type":{"text/html; charset=utf-8"}},Body:console.HTML},nil};return (api.API{Store:r.db}).Handle(req),nil
 }
 var req execution;if err:=json.Unmarshal(raw,&req);err!=nil{return nil,err}
 switch method{
 case "request.intercept_before":scope,_:=req.Metadata["caller_scope"].(string);model:=req.RequestedModel;if model==""{model=req.Model};return intercept(r.billing.Before(req.RequestID,scope,model,req.Body))
 case "request.intercept_after":return intercept(r.billing.After(req.RequestID,req.Model))
 case "response.intercept_after":return map[string]any{},r.billing.Response(req.RequestID,req.Body,false)
 case "response.intercept_stream_chunk":return map[string]any{},r.billing.Response(req.RequestID,req.Body,true)
 case "websocket.response_event":if req.EventType=="response.completed"||req.EventType=="response.done"{return map[string]any{},r.billing.Response(req.RequestID,req.Payload,false)};return map[string]any{},nil
 case "request.complete":return map[string]any{},r.billing.Complete(req.RequestID)
 }
 return nil,errors.New("未知插件方法: "+method)
}
func intercept(err error)(any,error){if err==nil{return map[string]any{},nil};code:=503;reason:="billing_unavailable";var denial *billing.Denial;if errors.As(err,&denial){code=denial.Status;reason=denial.Reason};return map[string]any{"Terminate":true,"StatusCode":code,"ResponseHeaders":http.Header{"Content-Type":{"application/json"}},"ResponseBody":[]byte(store.JSON(map[string]any{"error":map[string]any{"type":"cpa2pool_rejected","message":reason}}))},nil}

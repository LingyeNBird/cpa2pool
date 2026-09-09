package billing

import (
 "bytes"
 "encoding/json"
 "strings"
 "cpa2pool/internal/domain"
)

type Meter struct { Usage domain.Usage; Model string; Tier string; Seen bool; buffer []byte }
func object(v any)map[string]any{m,_:=v.(map[string]any);return m}
func number(m map[string]any,k string)int64{v,_:=m[k].(float64);return int64(v)}
func str(m map[string]any,k string)string{v,_:=m[k].(string);return v}
func assign(m map[string]any,key string,dst *int64){if _,ok:=m[key];ok{*dst=number(m,key)}}
func(m *Meter) JSON(body []byte){
 var root map[string]any;if json.Unmarshal(body,&root)!=nil{return}
 if response:=object(root["response"]);response!=nil{root=response}else if message:=object(root["message"]);message!=nil{root=message}
 if model:=str(root,"model");model!=""{m.Model=model};if tier:=str(root,"service_tier");tier!=""{m.Tier=tier}
 if u:=object(root["usage"]);u!=nil{
  m.Seen=true
  _,anthropicRead:=u["cache_read_input_tokens"];_,anthropicWrite:=u["cache_creation_input_tokens"]
  if anthropicRead||anthropicWrite||str(root,"type")=="message"||str(root,"type")=="message_delta"{
   assign(u,"input_tokens",&m.Usage.Input);assign(u,"output_tokens",&m.Usage.Output);assign(u,"cache_read_input_tokens",&m.Usage.CacheRead);assign(u,"cache_creation_input_tokens",&m.Usage.CacheWrite)
  }else{
   in,hasInput:=u["input_tokens"];if !hasInput{in,hasInput=u["prompt_tokens"]}
   details:=object(u["input_tokens_details"]);if details==nil{details=object(u["prompt_tokens_details"])}
   if hasInput{m.Usage.CacheRead=number(details,"cached_tokens");m.Usage.Input=int64(in.(float64))-m.Usage.CacheRead}
   assign(u,"output_tokens",&m.Usage.Output);assign(u,"completion_tokens",&m.Usage.Output)
   details=object(u["output_tokens_details"]);if details==nil{details=object(u["completion_tokens_details"])};assign(details,"reasoning_tokens",&m.Usage.Reasoning)
  }
 }
 if u:=object(root["usageMetadata"]);u!=nil{m.Seen=true;m.Usage.CacheRead=number(u,"cachedContentTokenCount");m.Usage.Input=number(u,"promptTokenCount")-m.Usage.CacheRead;m.Usage.Reasoning=number(u,"thoughtsTokenCount");m.Usage.Output=number(u,"candidatesTokenCount")+m.Usage.Reasoning}
}
func(m *Meter) Stream(chunk []byte){
 m.buffer=append(m.buffer,chunk...)
 for {i:=bytes.IndexByte(m.buffer,'\n');if i<0{break};line:=bytes.TrimSpace(m.buffer[:i]);if bytes.HasPrefix(line,[]byte("data:")){m.JSON(bytes.TrimSpace(line[5:]))};m.buffer=m.buffer[i+1:]}
}
func(m *Meter) Flush(){line:=bytes.TrimSpace(m.buffer);if bytes.HasPrefix(line,[]byte("data:")){m.JSON(bytes.TrimSpace(line[5:]))}else{m.JSON(line)};m.buffer=nil}
func RequestPolicy(body []byte,model string)(string,string,string){var root map[string]any;_ = json.Unmarshal(body,&root);effort:=str(root,"reasoning_effort");if effort==""{effort=str(object(root["reasoning"]),"effort")};if effort==""{effort=str(object(root["output_config"]),"effort")};if effort==""{effort=str(object(object(root["generationConfig"])["thinkingConfig"]),"thinkingLevel")};if i:=strings.LastIndex(model,"(");i>=0&&strings.HasSuffix(model,")"){if effort==""{effort=model[i+1:len(model)-1]};model=model[:i]};if effort==""{effort="default"};return model,strings.ToLower(effort),str(root,"service_tier")}

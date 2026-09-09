package billing

import "testing"

func TestAnthropicStreamAccumulatesInputAndFinalOutput(t *testing.T){
 var m Meter
 chunks:=[]string{"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"type\":\"message\",\"model\":\"claude-test\",\"usage\":{\"input_tokens\":10,\"cache_read_input_tokens\":20,\"cache_creation_input_tokens\":30,\"output_tokens\":1}}}\n\n", "data: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":", "50}}\n\n"}
 for _,c:=range chunks{m.Stream([]byte(c))};m.Flush()
 if m.Usage.Input!=10||m.Usage.Output!=50||m.Usage.CacheRead!=20||m.Usage.CacheWrite!=30{t.Fatalf("merged usage: %+v",m.Usage)}
}
func TestOpenAIReasoningIsNotBilledTwice(t *testing.T){var m Meter;m.JSON([]byte(`{"usage":{"input_tokens":100,"input_tokens_details":{"cached_tokens":40},"output_tokens":20,"output_tokens_details":{"reasoning_tokens":15}}}`));if m.Usage.Input!=60||m.Usage.CacheRead!=40||m.Usage.Output!=20||m.Usage.Reasoning!=15{t.Fatalf("usage: %+v",m.Usage)}}

package billing

import (
	"bytes"
	"cpa2pool/internal/domain"
	"encoding/json"
	"strconv"
	"strings"
)

type Meter struct {
	Usage  domain.Usage
	Model  string
	Tier   string
	Seen   bool
	buffer []byte
}

func object(v any) map[string]any             { m, _ := v.(map[string]any); return m }
func number(m map[string]any, k string) int64 { v, _ := m[k].(float64); return int64(v) }
func str(m map[string]any, k string) string   { v, _ := m[k].(string); return v }
func assign(m map[string]any, key string, dst *int64) {
	if _, ok := m[key]; ok {
		*dst = number(m, key)
	}
}
func (m *Meter) JSON(body []byte) {
	var root map[string]any
	if json.Unmarshal(body, &root) != nil {
		return
	}
	if response := object(root["response"]); response != nil {
		root = response
	} else if message := object(root["message"]); message != nil {
		root = message
	}
	if model := str(root, "model"); model != "" {
		m.Model = model
	}
	if tier := str(root, "service_tier"); tier != "" {
		m.Tier = tier
	}
	if u := object(root["usage"]); u != nil {
		m.Seen = true
		_, anthropicRead := u["cache_read_input_tokens"]
		_, anthropicWrite := u["cache_creation_input_tokens"]
		if anthropicRead || anthropicWrite || str(root, "type") == "message" || str(root, "type") == "message_delta" {
			assign(u, "input_tokens", &m.Usage.Input)
			assign(u, "output_tokens", &m.Usage.Output)
			assign(u, "cache_read_input_tokens", &m.Usage.CacheRead)
			assign(u, "cache_creation_input_tokens", &m.Usage.CacheWrite)
		} else {
			in, hasInput := u["input_tokens"]
			if !hasInput {
				in, hasInput = u["prompt_tokens"]
			}
			details := object(u["input_tokens_details"])
			if details == nil {
				details = object(u["prompt_tokens_details"])
			}
			if hasInput {
				m.Usage.CacheRead = number(details, "cached_tokens")
				m.Usage.Input = int64(in.(float64)) - m.Usage.CacheRead
			}
			assign(u, "output_tokens", &m.Usage.Output)
			assign(u, "completion_tokens", &m.Usage.Output)
			details = object(u["output_tokens_details"])
			if details == nil {
				details = object(u["completion_tokens_details"])
			}
			assign(details, "reasoning_tokens", &m.Usage.Reasoning)
		}
	}
	if u := object(root["usageMetadata"]); u != nil {
		m.Seen = true
		m.Usage.CacheRead = number(u, "cachedContentTokenCount")
		m.Usage.Input = number(u, "promptTokenCount") - m.Usage.CacheRead
		m.Usage.Reasoning = number(u, "thoughtsTokenCount")
		m.Usage.Output = number(u, "candidatesTokenCount") + m.Usage.Reasoning
	}
}
func (m *Meter) Stream(chunk []byte) {
	m.buffer = append(m.buffer, chunk...)
	for {
		i := bytes.IndexByte(m.buffer, '\n')
		if i < 0 {
			break
		}
		line := bytes.TrimSpace(m.buffer[:i])
		if bytes.HasPrefix(line, []byte("data:")) {
			m.JSON(bytes.TrimSpace(line[5:]))
		}
		m.buffer = m.buffer[i+1:]
	}
	// CPA can pass complete SSE data lines without trailing newlines.
	line := bytes.TrimSpace(m.buffer)
	payload := line
	if bytes.HasPrefix(line, []byte("data:")) {
		payload = bytes.TrimSpace(line[5:])
	}
	if json.Valid(payload) || bytes.Equal(payload, []byte("[DONE]")) {
		m.JSON(payload)
		m.buffer = nil
	}
}
func (m *Meter) Flush() {
	line := bytes.TrimSpace(m.buffer)
	if bytes.HasPrefix(line, []byte("data:")) {
		m.JSON(bytes.TrimSpace(line[5:]))
	} else {
		m.JSON(line)
	}
	m.buffer = nil
}
func imageSizeTier(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	switch value {
	case "1K", "2K", "4K":
		return value
	}
	parts := strings.Split(value, "X")
	if len(parts) == 2 {
		width, widthErr := strconv.Atoi(parts[0])
		height, heightErr := strconv.Atoi(parts[1])
		if widthErr == nil && heightErr == nil {
			edge := max(width, height)
			if edge <= 1024 {
				return "1K"
			}
			if edge <= 2048 {
				return "2K"
			}
			return "4K"
		}
	}
	return "2K"
}

func ImagePolicy(body []byte, model string) (bool, string, int64, string) {
	var root map[string]any
	_ = json.Unmarshal(body, &root)
	normalized := strings.ToLower(strings.TrimSpace(model))
	image := strings.HasPrefix(normalized, "gpt-image-") ||
		strings.HasPrefix(normalized, "grok-imagine-image") ||
		strings.HasPrefix(normalized, "grok-imagine-edit") ||
		strings.HasPrefix(normalized, "gemini-") && strings.Contains(normalized, "-image")
	size := str(root, "size")
	if config := object(object(root["generationConfig"])["imageConfig"]); config != nil {
		if value := str(config, "imageSize"); value != "" {
			size = value
		}
	}
	if tools, ok := root["tools"].([]any); ok {
		for _, raw := range tools {
			tool := object(raw)
			if str(tool, "type") != "image_generation" {
				continue
			}
			image = true
			if value := str(tool, "size"); value != "" {
				size = value
			}
			if value := str(tool, "model"); value != "" {
				model = value
			}
		}
	}
	count := number(root, "n")
	if count <= 0 {
		count = 1
	}
	return image, imageSizeTier(size), count, model
}

func ImageResponseCount(body []byte) int64 {
	var root map[string]any
	if json.Unmarshal(body, &root) != nil {
		return 0
	}
	if data, ok := root["data"].([]any); ok {
		return int64(len(data))
	}
	return 0
}

func VideoPolicy(body []byte, model string) (bool, int64, string) {
	var root map[string]any
	if json.Unmarshal(body, &root) != nil {
		return false, 0, ""
	}
	normalized := strings.ToLower(strings.TrimSpace(model))
	videoModel := normalized == "sora-2" ||
		strings.HasPrefix(normalized, "sora-2-") ||
		normalized == "grok-imagine-video" ||
		strings.HasPrefix(normalized, "grok-imagine-video-")
	if !videoModel {
		return false, 0, ""
	}
	if _, hasPrompt := root["prompt"]; !hasPrompt {
		return false, 0, ""
	}
	seconds := int64(4)
	switch value := root["seconds"].(type) {
	case string:
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			seconds = parsed
		}
	case float64:
		seconds = int64(value)
	}
	if seconds < 1 {
		seconds = 1
	}
	resolution := strings.ToLower(str(root, "resolution"))
	if resolution == "" {
		switch strings.ToLower(str(root, "size")) {
		case "1024x1792", "1792x1024":
			resolution = "1024p"
		case "1080x1920", "1920x1080":
			resolution = "1080p"
		case "720x1280", "1280x720":
			resolution = "720p"
		}
	}
	switch resolution {
	case "480p", "720p", "1024p", "1080p":
	default:
		resolution = "720p"
	}
	return true, seconds, resolution
}

func RequestPolicy(body []byte, model string) (string, string, string) {
	var root map[string]any
	_ = json.Unmarshal(body, &root)
	effort := str(root, "reasoning_effort")
	if effort == "" {
		effort = str(object(root["reasoning"]), "effort")
	}
	if effort == "" {
		effort = str(object(root["output_config"]), "effort")
	}
	if effort == "" {
		effort = str(object(object(root["generationConfig"])["thinkingConfig"]), "thinkingLevel")
	}
	if i := strings.LastIndex(model, "("); i >= 0 && strings.HasSuffix(model, ")") {
		if effort == "" {
			effort = model[i+1 : len(model)-1]
		}
		model = model[:i]
	}
	if effort == "" {
		effort = "default"
	}
	return model, strings.ToLower(effort), str(root, "service_tier")
}

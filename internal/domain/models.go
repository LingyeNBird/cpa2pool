package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

// Money is USD in billionths. API amounts use decimal USD strings.
type Money int64

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(`"` + decimal.New(int64(m), -9).String() + `"`), nil
}
func (m *Money) UnmarshalJSON(b []byte) error {
	d, err := decimal.NewFromString(stringTrimQuotes(b))
	if err != nil {
		return err
	}
	if !d.Equal(d.Round(9)) {
		return errors.New("金额最多保留九位小数")
	}
	*m = Money(d.Shift(9).IntPart())
	return nil
}
func stringTrimQuotes(b []byte) string {
	if len(b) > 1 && b[0] == '"' {
		return string(b[1 : len(b)-1])
	}
	return string(b)
}
func ID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Now() time.Time { return time.Now().UTC() }

type Participant struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Note       string     `json:"note"`
	KeyScope   string     `json:"key_scope"`
	KeyPreview string     `json:"key_preview"`
	Enabled    bool       `json:"enabled"`
	Paused     bool       `json:"paused"`
	Deleted    bool       `json:"deleted"`
	Models     []string   `json:"models"`
	Efforts    []string   `json:"efforts"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
type Quota struct {
	ID            string     `json:"id"`
	ParticipantID string     `json:"participant_id"`
	Name          string     `json:"name"`
	Limit         Money      `json:"limit"`
	Period        string     `json:"period"`
	StartsAt      time.Time  `json:"starts_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	Enabled       bool       `json:"enabled"`
	Anchor        time.Time  `json:"anchor"`
}
type Period struct {
	ID            string     `json:"id"`
	QuotaID       string     `json:"quota_id"`
	ParticipantID string     `json:"participant_id"`
	StartsAt      time.Time  `json:"starts_at"`
	EndsAt        *time.Time `json:"ends_at"`
	ClosedAt      *time.Time `json:"closed_at"`
	Limit         Money      `json:"limit"`
	Used          Money      `json:"used"`
	Remaining     Money      `json:"remaining"`
}
type QuotaView struct {
	Quota
	Current *Period `json:"current"`
	Reason  string  `json:"reason"`
}
type Status struct {
	Available bool        `json:"available"`
	Reasons   []string    `json:"reasons"`
	Remaining *Money      `json:"remaining"`
	Used      Money       `json:"used"`
	Quotas    []QuotaView `json:"quotas"`
}
type Usage struct {
	Input           int64  `json:"input"`
	Output          int64  `json:"output"`
	CacheRead       int64  `json:"cache_read"`
	CacheWrite      int64  `json:"cache_write"`
	Reasoning       int64  `json:"reasoning"`
	Images          int64  `json:"images"`
	ImageSize       string `json:"image_size"`
	VideoSeconds    int64  `json:"video_seconds"`
	VideoResolution string `json:"video_resolution"`
}
type Price struct {
	BillingMode          string          `json:"billing_mode"`
	ImagePrice1K         Money           `json:"image_price_1k"`
	ImagePrice2K         Money           `json:"image_price_2k"`
	ImagePrice4K         Money           `json:"image_price_4k"`
	VideoPrice480p       Money           `json:"video_price_480p"`
	VideoPrice720p       Money           `json:"video_price_720p"`
	VideoPrice1024p      Money           `json:"video_price_1024p"`
	VideoPrice1080p      Money           `json:"video_price_1080p"`
	Model                string          `json:"model"`
	Input                Money           `json:"input"`
	Output               Money           `json:"output"`
	CacheRead            Money           `json:"cache_read"`
	CacheWrite           Money           `json:"cache_write"`
	PriorityEnabled      bool            `json:"priority_enabled"`
	PriorityMultiplier   decimal.Decimal `json:"priority_multiplier"`
	LongEnabled          bool            `json:"long_enabled"`
	LongThreshold        int64           `json:"long_threshold"`
	LongInputMultiplier  decimal.Decimal `json:"long_input_multiplier"`
	LongOutputMultiplier decimal.Decimal `json:"long_output_multiplier"`
	ModelEnabled         bool            `json:"model_enabled"`
	ModelMultiplier      decimal.Decimal `json:"model_multiplier"`
	Combination          string          `json:"combination"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
type Factor struct {
	Name   string          `json:"name"`
	Input  decimal.Decimal `json:"input"`
	Output decimal.Decimal `json:"output"`
}
type Charge struct {
	Base    Money    `json:"base"`
	Final   Money    `json:"final"`
	Factors []Factor `json:"factors"`
	Price   Price    `json:"price"`
}
type Bill struct {
	ID             string    `json:"id"`
	RequestID      string    `json:"request_id"`
	ParticipantID  string    `json:"participant_id"`
	Model          string    `json:"model"`
	RequestedModel string    `json:"requested_model"`
	Effort         string    `json:"effort"`
	ServiceTier    string    `json:"service_tier"`
	Time           time.Time `json:"time"`
	Usage          Usage     `json:"usage"`
	Charge         Charge    `json:"charge"`
}
type Audit struct {
	ID            string    `json:"id"`
	ParticipantID string    `json:"participant_id"`
	QuotaID       string    `json:"quota_id"`
	Action        string    `json:"action"`
	Time          time.Time `json:"time"`
	Before        any       `json:"before"`
	After         any       `json:"after"`
	Note          string    `json:"note"`
}

package billing

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/participant"
	"cpa2pool/internal/pricing"
	"cpa2pool/internal/quota"
	"cpa2pool/internal/store"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sync"
)

type Pending struct {
	ParticipantID   string
	RequestedModel  string
	Model           string
	Effort          string
	Tier            string
	Prices          map[string]domain.Price
	Meter           Meter
	Charged         bool
	Image           bool
	ImageSize       string
	ImageCount      int64
	Video           bool
	VideoSeconds    int64
	VideoResolution string
}
type Service struct {
	Store   *store.Store
	mu      sync.Mutex
	pending map[string]*Pending
}

func New(db *store.Store) *Service { return &Service{Store: db, pending: map[string]*Pending{}} }

type Denial struct {
	Status int
	Reason string
}

func normalizePriceMode(price domain.Price) string {
	if price.BillingMode == "" {
		return "token"
	}
	return price.BillingMode
}

func (d *Denial) Error() string { return d.Reason }
func (s *Service) Before(id, scope, model string, body []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pending[id]; ok {
		return nil
	}
	requested, effort, tier := RequestPolicy(body, model)
	image, imageSize, imageCount, imageModel := ImagePolicy(body, requested)
	video, videoSeconds, videoResolution := VideoPolicy(body, requested)
	if image {
		requested = imageModel
	}
	var p domain.Participant
	err := s.Store.Tx(func(tx *sql.Tx) error {
		var e error
		p, e = participant.ByScope(tx, scope)
		if e == sql.ErrNoRows {
			return &Denial{403, "participant_not_registered"}
		}
		if e != nil {
			return e
		}
		status, e := quota.Status(tx, p, domain.Now())
		if e != nil {
			return e
		}
		if !status.Available {
			return &Denial{402, status.Reasons[0]}
		}
		if len(p.Models) > 0 && !slices.Contains(p.Models, requested) {
			return &Denial{403, "model_not_allowed"}
		}
		if len(p.Efforts) > 0 && !slices.Contains(p.Efforts, effort) {
			return &Denial{403, "reasoning_not_allowed"}
		}
		return nil
	})
	if err != nil {
		return err
	}
	prices, err := (pricing.Service{Store: s.Store}).List()
	if err != nil {
		return err
	}
	catalog := make(map[string]domain.Price, len(prices))
	for _, price := range prices {
		catalog[price.Model] = price
	}
	s.pending[id] = &Pending{
		ParticipantID: p.ID, RequestedModel: requested, Model: requested, Effort: effort, Tier: tier,
		Prices: catalog, Image: image, ImageSize: imageSize, ImageCount: imageCount,
		Video: video, VideoSeconds: videoSeconds, VideoResolution: videoResolution,
	}
	return nil
}
func (s *Service) After(id, model string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.pending[id]
	if p == nil {
		return &Denial{403, "request_not_admitted"}
	}
	if p.Image || p.Video {
		model = p.Model
	} else {
		model, _, _ = RequestPolicy(nil, model)
	}
	if _, ok := p.Prices[model]; !ok {
		return &Denial{403, "model_price_missing:" + model}
	}
	if p.Image && normalizePriceMode(p.Prices[model]) != "image" {
		return &Denial{403, "image_price_missing:" + model}
	}
	if p.Video && pricing.VideoUnitPrice(p.Prices[model], p.VideoResolution) == 0 {
		return &Denial{403, "video_resolution_price_missing:" + model + ":" + p.VideoResolution}
	}
	p.Model = model
	return nil
}
func (s *Service) Response(id string, body []byte, stream bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.pending[id]
	if p == nil || p.Charged {
		return nil
	}
	if stream {
		p.Meter.Stream(body)
		return nil
	}
	p.Meter.JSON(body)
	if p.Image {
		if count := ImageResponseCount(body); count > 0 {
			p.ImageCount = count
		}
	}
	return s.charge(id, p)
}
func (s *Service) Complete(id string, succeeded bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.pending[id]
	if p == nil {
		return nil
	}
	defer delete(s.pending, id)
	p.Meter.Flush()
	if p.Image || p.Video {
		if !succeeded || p.Charged {
			return nil
		}
		return s.charge(id, p)
	}
	if !p.Meter.Seen || p.Charged {
		return nil
	}
	return s.charge(id, p)
}
func (s *Service) charge(id string, p *Pending) error {
	if !p.Image && !p.Video && !p.Meter.Seen {
		return errors.New("上游未返回 Token 用量，无法计费")
	}
	model := p.Model
	if !p.Image && !p.Video && p.Meter.Model != "" {
		model = p.Meter.Model
	}
	price, ok := p.Prices[model]
	if !ok {
		return fmt.Errorf("实际模型未配置价格: %s", model)
	}
	tier := p.Tier
	if p.Meter.Tier != "" {
		tier = p.Meter.Tier
	}
	usage := p.Meter.Usage
	charge := pricing.Calculate(price, usage, tier)
	if p.Image {
		usage.Images = p.ImageCount
		usage.ImageSize = p.ImageSize
		imageCharge := pricing.CalculateImage(price, p.ImageCount, p.ImageSize)
		charge.Base += imageCharge.Base
		charge.Final += imageCharge.Final
	}
	if p.Video {
		usage.VideoSeconds = p.VideoSeconds
		usage.VideoResolution = p.VideoResolution
		videoCharge := pricing.CalculateVideo(price, p.VideoSeconds, p.VideoResolution)
		charge.Base += videoCharge.Base
		charge.Final += videoCharge.Final
	}
	bill := domain.Bill{ID: domain.ID(), RequestID: id, ParticipantID: p.ParticipantID, Model: model, RequestedModel: p.RequestedModel, Effort: p.Effort, ServiceTier: tier, Time: domain.Now(), Usage: usage, Charge: charge}
	err := s.Store.Tx(func(tx *sql.Tx) error {
		result, e := tx.Exec("INSERT INTO bills VALUES(?,?,?,?,?,?,?) ON CONFLICT(request_id) DO NOTHING", bill.ID, id, bill.ParticipantID, model, bill.Time.Format("2006-01-02T15:04:05.000000000Z"), int64(bill.Charge.Final), store.JSON(bill))
		if e != nil {
			return e
		}
		n, e := result.RowsAffected()
		if e != nil || n == 0 {
			return e
		}
		qs, e := quota.List(tx, p.ParticipantID)
		if e != nil {
			return e
		}
		for _, q := range qs {
			if !quota.Active(q, bill.Time) {
				continue
			}
			period, e := quota.Current(tx, q, bill.Time)
			if e != nil {
				return e
			}
			if _, e = tx.Exec("UPDATE periods SET used=used+? WHERE id=?", int64(bill.Charge.Final), period.ID); e != nil {
				return e
			}
		}
		return nil
	})
	if err == nil {
		p.Charged = true
	}
	return err
}

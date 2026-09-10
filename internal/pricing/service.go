package pricing

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/store"
	"database/sql"
	"errors"
	"github.com/shopspring/decimal"
)

type Service struct{ Store *store.Store }

func (s Service) List() ([]domain.Price, error) {
	return store.List[domain.Price](s.Store.DB, "SELECT body FROM prices ORDER BY model")
}
func Get(q store.Query, model string) (domain.Price, error) {
	return store.One[domain.Price](q, "SELECT body FROM prices WHERE model=?", model)
}
func normalize(p domain.Price) domain.Price {
	if p.BillingMode == "" {
		p.BillingMode = "token"
	}
	return p
}

func validate(p domain.Price) error {
	if p.Model == "" {
		return errors.New("请输入模型名称")
	}
	if p.Input < 0 || p.Output < 0 || p.CacheRead < 0 || p.CacheWrite < 0 ||
		p.ImagePrice1K < 0 || p.ImagePrice2K < 0 || p.ImagePrice4K < 0 ||
		p.VideoPrice480p < 0 || p.VideoPrice720p < 0 || p.VideoPrice1024p < 0 || p.VideoPrice1080p < 0 {
		return errors.New("单价不能为负")
	}
	if p.BillingMode != "token" && p.BillingMode != "image" {
		return errors.New("计费模式应为 token 或 image")
	}
	if p.Combination != "multiply" && p.Combination != "max" {
		return errors.New("倍率组合应为 multiply 或 max")
	}
	if p.PriorityEnabled && !p.PriorityMultiplier.IsPositive() || p.LongEnabled && (p.LongThreshold < 0 || !p.LongInputMultiplier.IsPositive() || !p.LongOutputMultiplier.IsPositive()) || p.ModelEnabled && !p.ModelMultiplier.IsPositive() {
		return errors.New("启用的倍率必须大于零")
	}
	return nil
}

func (s Service) Save(p domain.Price) (domain.Price, error) {
	p = normalize(p)
	if err := validate(p); err != nil {
		return p, err
	}
	p.UpdatedAt = domain.Now()
	err := s.Store.Tx(func(tx *sql.Tx) error {
		old, e := Get(tx, p.Model)
		if e != nil && e != sql.ErrNoRows {
			return e
		}
		_, e = tx.Exec("INSERT INTO prices VALUES(?,?) ON CONFLICT(model) DO UPDATE SET body=excluded.body", p.Model, store.JSON(p))
		if e != nil {
			return e
		}
		return store.Audit(tx, "", "", "price.save", "", old, p)
	})
	return p, err
}
func (s Service) SyncDefaults(defaults []domain.Price) ([]domain.Price, error) {
	now := domain.Now()
	changed := make([]domain.Price, 0, len(defaults))
	err := s.Store.Tx(func(tx *sql.Tx) error {
		for _, p := range defaults {
			p = normalize(p)
			if err := validate(p); err != nil {
				return err
			}
			p.UpdatedAt = now
			old, err := Get(tx, p.Model)
			if err == nil {
				updated := false
				if p.BillingMode == "image" && old.BillingMode == "" &&
					old.ImagePrice1K == 0 && old.ImagePrice2K == 0 && old.ImagePrice4K == 0 {
					old.BillingMode = "image"
					old.ImagePrice1K = p.ImagePrice1K
					old.ImagePrice2K = p.ImagePrice2K
					old.ImagePrice4K = p.ImagePrice4K
					updated = true
				}
				if old.VideoPrice480p == 0 && old.VideoPrice720p == 0 &&
					old.VideoPrice1024p == 0 && old.VideoPrice1080p == 0 &&
					(p.VideoPrice480p != 0 || p.VideoPrice720p != 0 ||
						p.VideoPrice1024p != 0 || p.VideoPrice1080p != 0) {
					old.VideoPrice480p = p.VideoPrice480p
					old.VideoPrice720p = p.VideoPrice720p
					old.VideoPrice1024p = p.VideoPrice1024p
					old.VideoPrice1080p = p.VideoPrice1080p
					updated = true
				}
				if !updated {
					continue
				}
				old.UpdatedAt = now
				if _, err = tx.Exec("UPDATE prices SET body=? WHERE model=?", store.JSON(old), old.Model); err != nil {
					return err
				}
				changed = append(changed, old)
				continue
			}
			if err != sql.ErrNoRows {
				return err
			}
			if _, err = tx.Exec("INSERT INTO prices VALUES(?,?)", p.Model, store.JSON(p)); err != nil {
				return err
			}
			changed = append(changed, p)
		}
		if len(changed) == 0 {
			return nil
		}
		return store.Audit(tx, "", "", "price.sync_defaults", "", nil, changed)
	})
	if err != nil {
		return nil, err
	}
	return s.List()
}

func (s Service) Delete(model string) error {
	return s.Store.Tx(func(tx *sql.Tx) error {
		p, e := Get(tx, model)
		if e != nil {
			return e
		}
		if _, e = tx.Exec("DELETE FROM prices WHERE model=?", model); e != nil {
			return e
		}
		return store.Audit(tx, "", "", "price.delete", "", p, nil)
	})
}

// All rates are USD/million tokens. Round exactly once, after applying factors.
func Calculate(p domain.Price, u domain.Usage, tier string) domain.Charge {
	one := decimal.NewFromInt(1)
	factors := []domain.Factor{}
	if p.PriorityEnabled && (tier == "priority" || tier == "fast") {
		factors = append(factors, domain.Factor{Name: "priority", Input: p.PriorityMultiplier, Output: p.PriorityMultiplier})
	}
	if p.LongEnabled && u.Input+u.CacheRead+u.CacheWrite > p.LongThreshold {
		factors = append(factors, domain.Factor{Name: "long_context", Input: p.LongInputMultiplier, Output: p.LongOutputMultiplier})
	}
	if p.ModelEnabled {
		factors = append(factors, domain.Factor{Name: "model", Input: p.ModelMultiplier, Output: p.ModelMultiplier})
	}
	im, om := one, one
	for i, f := range factors {
		if p.Combination == "max" {
			if i == 0 {
				im, om = f.Input, f.Output
			} else {
				im = decimal.Max(im, f.Input)
				om = decimal.Max(om, f.Output)
			}
		} else {
			im = im.Mul(f.Input)
			om = om.Mul(f.Output)
		}
	}
	tokenCost := func(n int64, r domain.Money) decimal.Decimal {
		return decimal.NewFromInt(n).Mul(decimal.NewFromInt(int64(r))).Div(decimal.NewFromInt(1000000))
	}
	input := tokenCost(u.Input, p.Input).Add(tokenCost(u.CacheRead, p.CacheRead)).Add(tokenCost(u.CacheWrite, p.CacheWrite))
	output := tokenCost(u.Output, p.Output)
	return domain.Charge{Base: domain.Money(input.Add(output).Round(0).IntPart()), Final: domain.Money(input.Mul(im).Add(output.Mul(om)).Round(0).IntPart()), Factors: factors, Price: p}
}

func CalculateImage(p domain.Price, count int64, size string) domain.Charge {
	unit := p.ImagePrice2K
	switch size {
	case "1K":
		unit = p.ImagePrice1K
	case "4K":
		unit = p.ImagePrice4K
	}
	base := decimal.NewFromInt(int64(unit)).Mul(decimal.NewFromInt(count))
	final := base
	factors := []domain.Factor{}
	if p.ModelEnabled {
		final = final.Mul(p.ModelMultiplier)
		factors = append(factors, domain.Factor{Name: "model", Input: p.ModelMultiplier, Output: p.ModelMultiplier})
	}
	return domain.Charge{
		Base:    domain.Money(base.Round(0).IntPart()),
		Final:   domain.Money(final.Round(0).IntPart()),
		Factors: factors,
		Price:   p,
	}
}

func VideoUnitPrice(p domain.Price, resolution string) domain.Money {
	switch resolution {
	case "480p":
		return p.VideoPrice480p
	case "1024p":
		return p.VideoPrice1024p
	case "1080p":
		return p.VideoPrice1080p
	default:
		return p.VideoPrice720p
	}
}

func CalculateVideo(p domain.Price, seconds int64, resolution string) domain.Charge {
	unit := VideoUnitPrice(p, resolution)
	base := decimal.NewFromInt(int64(unit)).Mul(decimal.NewFromInt(seconds))
	final := base
	factors := []domain.Factor{}
	if p.ModelEnabled {
		final = final.Mul(p.ModelMultiplier)
		factors = append(factors, domain.Factor{Name: "model", Input: p.ModelMultiplier, Output: p.ModelMultiplier})
	}
	return domain.Charge{
		Base:    domain.Money(base.Round(0).IntPart()),
		Final:   domain.Money(final.Round(0).IntPart()),
		Factors: factors,
		Price:   p,
	}
}

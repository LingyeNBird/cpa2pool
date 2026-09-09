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
func (s Service) Save(p domain.Price) (domain.Price, error) {
	if p.Model == "" {
		return p, errors.New("请输入模型名称")
	}
	if p.Input < 0 || p.Output < 0 || p.CacheRead < 0 || p.CacheWrite < 0 {
		return p, errors.New("单价不能为负")
	}
	if p.Combination != "multiply" && p.Combination != "max" {
		return p, errors.New("倍率组合应为 multiply 或 max")
	}
	if p.PriorityEnabled && !p.PriorityMultiplier.IsPositive() || p.LongEnabled && (p.LongThreshold < 0 || !p.LongInputMultiplier.IsPositive() || !p.LongOutputMultiplier.IsPositive()) || p.ModelEnabled && !p.ModelMultiplier.IsPositive() {
		return p, errors.New("启用的倍率必须大于零")
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

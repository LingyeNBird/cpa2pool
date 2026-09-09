package quota

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/participant"
	"cpa2pool/internal/store"
	"database/sql"
	"errors"
	"time"
)

type Service struct{ Store *store.Store }
type Adjustment struct {
	QuotaID   string       `json:"quota_id"`
	Action    string       `json:"action"`
	Amount    domain.Money `json:"amount"`
	Restart   bool         `json:"restart"`
	ExpiresAt *time.Time   `json:"expires_at"`
	Note      string       `json:"note"`
}

func (s Service) Save(q domain.Quota) (domain.Quota, error) {
	now := domain.Now()
	if q.StartsAt.IsZero() {
		q.StartsAt = now
	}
	q.StartsAt = q.StartsAt.UTC()
	if err := valid(q); err != nil {
		return q, err
	}
	err := s.Store.Tx(func(tx *sql.Tx) error {
		if _, err := participant.Get(tx, q.ParticipantID); err != nil {
			return err
		}
		if q.ID == "" {
			q.ID = domain.ID()
			q.Anchor = q.StartsAt
			if _, err := tx.Exec("INSERT INTO quotas VALUES(?,?,?)", q.ID, q.ParticipantID, store.JSON(q)); err != nil {
				return err
			}
			if _, err := createPeriod(tx, q, q.StartsAt); err != nil {
				return err
			}
			return store.Audit(tx, q.ParticipantID, q.ID, "quota.create", "", nil, q)
		}
		old, err := Get(tx, q.ID)
		if err != nil {
			return err
		}
		if old.ParticipantID != q.ParticipantID || old.Period != q.Period || !old.StartsAt.Equal(q.StartsAt) {
			return errors.New("现有额度不能更改归属或周期类型；请创建另一项额度")
		}
		q.Anchor = old.Anchor
		p, err := Current(tx, old, now)
		if err != nil {
			return err
		}
		if p.ClosedAt != nil && Active(q, now) {
			p, err = createPeriod(tx, q, now)
			if err != nil {
				return err
			}
		}
		end := nextBoundary(q, p.StartsAt)
		if q.ExpiresAt != nil && (end == nil || q.ExpiresAt.Before(*end)) {
			end = q.ExpiresAt
		}
		if _, err = tx.Exec("UPDATE periods SET limit_amount=?,ends_at=? WHERE id=?", int64(q.Limit), nullable(end), p.ID); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE quotas SET body=? WHERE id=?", store.JSON(q), q.ID); err != nil {
			return err
		}
		return store.Audit(tx, q.ParticipantID, q.ID, "quota.save", "", old, q)
	})
	return q, err
}
func (s Service) Adjust(a Adjustment) (domain.Quota, error) {
	var out domain.Quota
	err := s.Store.Tx(func(tx *sql.Tx) error {
		q, err := Get(tx, a.QuotaID)
		if err != nil {
			return err
		}
		now := domain.Now()
		p, err := Current(tx, q, now)
		if err != nil {
			return err
		}
		before := struct {
			Quota  domain.Quota  `json:"quota"`
			Period domain.Period `json:"period"`
		}{q, p}
		switch a.Action {
		case "set_limit":
			q.Limit = a.Amount
			p.Limit = a.Amount
		case "add":
			q.Limit += a.Amount
			p.Limit += a.Amount
		case "set_remaining":
			if a.Amount < 0 {
				return errors.New("剩余额度不能为负")
			}
			q.Limit = p.Used + a.Amount
			p.Limit = q.Limit
		case "expiry":
			q.ExpiresAt = a.ExpiresAt
			p.EndsAt = nextBoundary(q, p.StartsAt)
			if q.ExpiresAt != nil && (p.EndsAt == nil || q.ExpiresAt.Before(*p.EndsAt)) {
				p.EndsAt = q.ExpiresAt
			}
		case "reset", "new_period":
			if _, err = tx.Exec("UPDATE periods SET closed_at=? WHERE id=?", text(now), p.ID); err != nil {
				return err
			}
			start, end := p.StartsAt, p.EndsAt
			if a.Restart || a.Action == "new_period" {
				q.Anchor = now
				start = now
			}
			p, err = createPeriod(tx, q, start)
			if err != nil {
				return err
			}
			if !a.Restart && a.Action == "reset" {
				p.EndsAt = end
			}
		default:
			return errors.New("未知额度操作")
		}
		if err = valid(q); err != nil {
			return err
		}
		if p.ClosedAt != nil && Active(q, now) && a.Action == "expiry" {
			p, err = createPeriod(tx, q, now)
			if err != nil {
				return err
			}
		}
		if _, err = tx.Exec("UPDATE periods SET limit_amount=?,used=?,ends_at=? WHERE id=?", int64(p.Limit), int64(p.Used), nullable(p.EndsAt), p.ID); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE quotas SET body=? WHERE id=?", store.JSON(q), q.ID); err != nil {
			return err
		}
		out = q
		p.Remaining = p.Limit - p.Used
		return store.Audit(tx, q.ParticipantID, q.ID, "quota."+a.Action, a.Note, before, struct {
			Quota  domain.Quota  `json:"quota"`
			Period domain.Period `json:"period"`
		}{q, p})
	})
	return out, err
}
func (s Service) Status(id string) (domain.Status, error) {
	var status domain.Status
	err := s.Store.Tx(func(tx *sql.Tx) error {
		p, err := participant.Get(tx, id)
		if err != nil {
			return err
		}
		status, err = Status(tx, p, domain.Now())
		return err
	})
	return status, err
}
func Status(tx *sql.Tx, p domain.Participant, now time.Time) (domain.Status, error) {
	s := domain.Status{Available: true, Reasons: []string{}, Quotas: []domain.QuotaView{}}
	if p.Deleted {
		s.Reasons = append(s.Reasons, "deleted")
	}
	if !p.Enabled {
		s.Reasons = append(s.Reasons, "disabled")
	}
	if p.Paused {
		s.Reasons = append(s.Reasons, "paused")
	}
	if p.ExpiresAt != nil && !now.Before(*p.ExpiresAt) {
		s.Reasons = append(s.Reasons, "expired")
	}
	qs, err := List(tx, p.ID)
	if err != nil {
		return s, err
	}
	active := 0
	for _, q := range qs {
		current, e := Current(tx, q, now)
		if e != nil {
			return s, e
		}
		v := domain.QuotaView{Quota: q, Current: &current}
		if !q.Enabled {
			v.Reason = "disabled"
		} else if now.Before(q.StartsAt) {
			v.Reason = "not_started"
		} else if q.ExpiresAt != nil && !now.Before(*q.ExpiresAt) {
			v.Reason = "expired"
		} else {
			active++
			if current.Remaining <= 0 {
				v.Reason = "exhausted"
				s.Reasons = append(s.Reasons, "quota_exhausted:"+q.ID)
			}
			if s.Remaining == nil || current.Remaining < *s.Remaining {
				r := current.Remaining
				s.Remaining = &r
			}
		}
		s.Quotas = append(s.Quotas, v)
	}
	if active == 0 {
		s.Reasons = append(s.Reasons, "no_active_quota")
	}
	s.Available = len(s.Reasons) == 0
	err = tx.QueryRow("SELECT COALESCE(SUM(cost),0) FROM bills WHERE participant_id=?", p.ID).Scan(&s.Used)
	return s, err
}

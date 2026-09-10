package api

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/participant"
	"cpa2pool/internal/pricing"
	"cpa2pool/internal/quota"
	"cpa2pool/internal/reporting"
	"cpa2pool/internal/store"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const Prefix = "/v0/management/cpa2pool"

type Request struct {
	Method  string
	Path    string
	Headers http.Header
	Query   url.Values
	Body    []byte
}
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}
type Route struct {
	Method string
	Path   string
}

func Routes() []Route {
	out := []Route{}
	for _, path := range []string{"participants", "prices", "quotas"} {
		for _, method := range []string{"GET", "POST", "PUT"} {
			out = append(out, Route{method, Prefix + "/" + path})
		}
	}
	for _, path := range []string{"participants", "prices"} {
		out = append(out, Route{"DELETE", Prefix + "/" + path})
	}
	out = append(out, Route{"POST", Prefix + "/adjustments"})
	out = append(out, Route{"POST", Prefix + "/participant-key-candidates"})
	out = append(out, Route{"POST", Prefix + "/prices/sync-defaults"})
	for _, path := range []string{"status", "bills", "stats", "periods", "audits"} {
		out = append(out, Route{"GET", Prefix + "/" + path})
	}
	return out
}

type API struct{ Store *store.Store }

func respond(code int, v any) Response {
	return Response{code, http.Header{"Content-Type": {"application/json; charset=utf-8"}, "Cache-Control": {"no-store"}}, []byte(store.JSON(v))}
}
func (a API) Handle(r Request) Response {
	data, err := a.dispatch(r)
	if err != nil {
		code := 400
		if errors.Is(err, sql.ErrNoRows) {
			code = 404
		}
		return respond(code, map[string]any{"error": err.Error()})
	}
	return respond(200, map[string]any{"data": data})
}
func decode[T any](b []byte) (T, error) { var v T; err := json.Unmarshal(b, &v); return v, err }
func decodePricePayload[T any](b []byte) (T, error) {
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		var zero T
		return zero, err
	}
	var strip func(any)
	strip = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			delete(value, "updated_at")
		case []any:
			for _, item := range value {
				strip(item)
			}
		}
	}
	strip(raw)
	clean, err := json.Marshal(raw)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](clean)
}

func filter(q url.Values) (reporting.Filter, error) {
	f := reporting.Filter{ParticipantID: q.Get("participant_id"), Model: q.Get("model"), Limit: 50}
	for _, name := range []string{"from", "to"} {
		if v := q.Get(name); v != "" {
			t, e := time.Parse(time.RFC3339Nano, v)
			if e != nil {
				return f, errors.New("时间应为 RFC3339 格式")
			}
			s := t.UTC().Format("2006-01-02T15:04:05.000000000Z")
			if name == "from" {
				f.From = s
			} else {
				f.To = s
			}
		}
	}
	if v := q.Get("limit"); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 1 || n > 200 {
			return f, errors.New("limit 应为 1 至 200")
		}
		f.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 {
			return f, errors.New("offset 不能为负")
		}
		f.Offset = n
	}
	return f, nil
}
func (a API) dispatch(r Request) (any, error) {
	ps := participant.Service{Store: a.Store}
	qs := quota.Service{Store: a.Store}
	prices := pricing.Service{Store: a.Store}
	reports := reporting.Service{Store: a.Store}
	path := strings.TrimPrefix(r.Path, Prefix)
	switch r.Method + " " + path {
	case "GET /participants":
		if id := r.Query.Get("id"); id != "" {
			return ps.Get(id)
		}
		return ps.List()
	case "POST /participant-key-candidates":
		v, e := decode[struct {
			Keys      []string `json:"keys"`
			CurrentID string   `json:"current_id"`
		}](r.Body)
		if e != nil {
			return nil, e
		}
		return ps.AvailableKeys(v.Keys, v.CurrentID)
	case "POST /participants", "PUT /participants":
		v, e := decode[participant.Input](r.Body)
		if e != nil {
			return nil, e
		}
		return ps.Save(v)
	case "DELETE /participants":
		return nil, ps.Delete(r.Query.Get("id"))
	case "GET /prices":
		return prices.List()
	case "POST /prices/sync-defaults":
		v, e := decodePricePayload[[]domain.Price](r.Body)
		if e != nil {
			return nil, e
		}
		return prices.SyncDefaults(v)
	case "POST /prices", "PUT /prices":
		v, e := decodePricePayload[domain.Price](r.Body)
		if e != nil {
			return nil, e
		}
		return prices.Save(v)
	case "DELETE /prices":
		return nil, prices.Delete(r.Query.Get("model"))
	case "GET /quotas":
		return quota.List(a.Store.DB, r.Query.Get("participant_id"))
	case "POST /quotas", "PUT /quotas":
		v, e := decode[domain.Quota](r.Body)
		if e != nil {
			return nil, e
		}
		return qs.Save(v)
	case "POST /adjustments":
		v, e := decode[quota.Adjustment](r.Body)
		if e != nil {
			return nil, e
		}
		return qs.Adjust(v)
	case "GET /status":
		return qs.Status(r.Query.Get("participant_id"))
	case "GET /periods":
		pid := r.Query.Get("participant_id")
		if pid != "" {
			if _, e := qs.Status(pid); e != nil {
				return nil, e
			}
		}
		return reports.Periods(pid, r.Query.Get("quota_id"))
	case "GET /bills", "GET /stats", "GET /audits":
		f, e := filter(r.Query)
		if e != nil {
			return nil, e
		}
		switch path {
		case "/bills":
			return reports.Bills(f)
		case "/stats":
			return reports.Stats(f, r.Query.Get("group"))
		default:
			return reports.Audits(f)
		}
	}
	return nil, sql.ErrNoRows
}

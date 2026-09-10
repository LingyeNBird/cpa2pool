package api

import (
	"cpa2pool/internal/pricing"
	"cpa2pool/internal/store"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
)

func TestSyncDefaultPricesIgnoresClientUpdatedAt(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Close()

	response := (API{Store: db}).Handle(Request{
		Method: http.MethodPost,
		Path:   Prefix + "/prices/sync-defaults",
		Query:  url.Values{},
		Body: []byte(`[{
			"model":"gpt-5.6-sol-wm",
			"input":"5","output":"30","cache_read":"0.5","cache_write":"6.25",
			"priority_enabled":true,"priority_multiplier":"2",
			"long_enabled":true,"long_threshold":272000,
			"long_input_multiplier":"2","long_output_multiplier":"1.5",
			"model_enabled":false,"model_multiplier":"1","combination":"multiply",
			"updated_at":""
		}]`),
	})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Handle() status = %d, body = %s", response.StatusCode, response.Body)
	}
	price, err := pricing.Get(db.DB, "gpt-5.6-sol-wm")
	if err != nil {
		t.Fatal(err)
	}
	if price.Input != 5_000_000_000 || price.Output != 30_000_000_000 {
		t.Fatalf("persisted price = input %d output %d", price.Input, price.Output)
	}
	if price.UpdatedAt.IsZero() {
		t.Fatal("server did not assign updated_at")
	}
}

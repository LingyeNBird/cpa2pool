package billing

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/participant"
	"cpa2pool/internal/pricing"
	"cpa2pool/internal/quota"
	"cpa2pool/internal/store"
	"path/filepath"
	"testing"
)

func TestImageRequestChargesPerSuccessfulImageExactlyOnce(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Close()
	key := "image-test-key"
	person, err := (participant.Service{Store: db}).Save(participant.Input{
		Participant: domain.Participant{Name: "image user", Enabled: true},
		APIKey:      key,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = (quota.Service{Store: db}).Save(domain.Quota{
		ParticipantID: person.ID,
		Name:          "image quota",
		Limit:         10_000_000_000,
		Period:        "none",
		Enabled:       true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = (pricing.Service{Store: db}).Save(domain.Price{
		Model:        "gpt-image-1",
		BillingMode:  "image",
		ImagePrice1K: 134_000_000,
		ImagePrice2K: 201_000_000,
		ImagePrice4K: 268_000_000,
		Combination:  "multiply",
	}); err != nil {
		t.Fatal(err)
	}

	service := New(db)
	if err = service.Before("image-1", participant.Scope(key), "gpt-5", []byte(`{"model":"gpt-5","tools":[{"type":"image_generation","model":"gpt-image-1","size":"1024x1024"}]}`)); err != nil {
		t.Fatal(err)
	}
	if err = service.After("image-1", "gpt-5"); err != nil {
		t.Fatal(err)
	}
	response := []byte(`{"data":[{"b64_json":"a"},{"b64_json":"b"}]}`)
	if err = service.Response("image-1", response, false); err != nil {
		t.Fatal(err)
	}
	if err = service.Response("image-1", response, false); err != nil {
		t.Fatal(err)
	}
	if err = service.Complete("image-1", true); err != nil {
		t.Fatal(err)
	}

	var billCount int
	var cost int64
	if err = db.DB.QueryRow("SELECT count(*),sum(cost) FROM bills").Scan(&billCount, &cost); err != nil {
		t.Fatal(err)
	}
	if billCount != 1 || cost != 268_000_000 {
		t.Fatalf("image bills = %d cost = %d, want 1 and 268000000", billCount, cost)
	}
}

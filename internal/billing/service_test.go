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
		Input:        1_000_000_000,
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
	response := []byte(`{"usage":{"input_tokens":1000,"output_tokens":0},"data":[{"b64_json":"a"},{"b64_json":"b"}]}`)
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
	if billCount != 1 || cost != 269_000_000 {
		t.Fatalf("combined image bill = %d cost = %d, want 1 and 269000000", billCount, cost)
	}
}

func TestVideoRequestChargesSecondsOnceAndIgnoresRetrieval(t *testing.T) {
	if video, _, _ := VideoPolicy([]byte(`{"request_id":"video_123"}`), "sora-2"); video {
		t.Fatal("video retrieval must not be classified as a generation")
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Close()
	key := "video-test-key"
	person, err := (participant.Service{Store: db}).Save(participant.Input{
		Participant: domain.Participant{Name: "video user", Enabled: true},
		APIKey:      key,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = (quota.Service{Store: db}).Save(domain.Quota{
		ParticipantID: person.ID,
		Name:          "video quota",
		Limit:         10_000_000_000,
		Period:        "none",
		Enabled:       true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = (pricing.Service{Store: db}).Save(domain.Price{
		Model:          "sora-2",
		VideoPrice720p: 100_000_000,
		Combination:    "multiply",
	}); err != nil {
		t.Fatal(err)
	}
	service := New(db)
	body := []byte(`{"model":"sora-2","prompt":"a bird in flight","seconds":"6","resolution":"720p"}`)
	if err = service.Before("video-1", participant.Scope(key), "sora-2", body); err != nil {
		t.Fatal(err)
	}
	if err = service.After("video-1", "sora-2"); err != nil {
		t.Fatal(err)
	}
	if err = service.Response("video-1", []byte(`{"id":"video_123","status":"queued"}`), false); err != nil {
		t.Fatal(err)
	}
	if err = service.Complete("video-1", true); err != nil {
		t.Fatal(err)
	}
	var billCount int
	var cost int64
	if err = db.DB.QueryRow("SELECT count(*),sum(cost) FROM bills").Scan(&billCount, &cost); err != nil {
		t.Fatal(err)
	}
	if billCount != 1 || cost != 600_000_000 {
		t.Fatalf("video bills = %d cost = %d, want 1 and 600000000", billCount, cost)
	}
}

package pricing

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/store"
	"path/filepath"
	"testing"
)

func TestSyncDefaultsAddsMissingPricesWithoutOverwritingSavedPrices(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Close()
	service := Service{Store: db}
	saved := domain.Price{Model: "custom", Input: 9, Output: 19, Combination: "multiply"}
	if _, err = service.Save(saved); err != nil {
		t.Fatal(err)
	}

	prices, err := service.SyncDefaults([]domain.Price{
		{Model: "custom", Input: 1, Output: 2, Combination: "multiply"},
		{Model: "discovered", Input: 3, Output: 4, Combination: "multiply"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 2 {
		t.Fatalf("SyncDefaults() returned %d prices, want 2", len(prices))
	}
	custom, err := Get(db.DB, "custom")
	if err != nil {
		t.Fatal(err)
	}
	if custom.Input != 9 || custom.Output != 19 {
		t.Fatalf("saved price was overwritten: input=%v output=%v", custom.Input, custom.Output)
	}
	discovered, err := Get(db.DB, "discovered")
	if err != nil {
		t.Fatal(err)
	}
	if discovered.Input != 3 || discovered.Output != 4 {
		t.Fatalf("discovered price not persisted: input=%v output=%v", discovered.Input, discovered.Output)
	}
}

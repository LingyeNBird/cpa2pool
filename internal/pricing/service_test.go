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
	legacyImage := domain.Price{Model: "legacy-image", Input: 7, Output: 8, Combination: "multiply"}
	if _, err = db.DB.Exec("INSERT INTO prices VALUES(?,?)", legacyImage.Model, store.JSON(legacyImage)); err != nil {
		t.Fatal(err)
	}

	prices, err := service.SyncDefaults([]domain.Price{
		{Model: "custom", Input: 1, Output: 2, Combination: "multiply"},
		{Model: "discovered", Input: 3, Output: 4, Combination: "multiply"},
		{
			Model: "legacy-image", BillingMode: "image",
			ImagePrice1K: 134_000_000, ImagePrice2K: 201_000_000, ImagePrice4K: 268_000_000,
			Combination: "multiply",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 3 {
		t.Fatalf("SyncDefaults() returned %d prices, want 3", len(prices))
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
	upgraded, err := Get(db.DB, "legacy-image")
	if err != nil {
		t.Fatal(err)
	}
	if upgraded.BillingMode != "image" || upgraded.ImagePrice2K != 201_000_000 {
		t.Fatalf("legacy image price was not upgraded: %+v", upgraded)
	}
	if upgraded.Input != 7 || upgraded.Output != 8 {
		t.Fatalf("legacy token prices were overwritten: input=%v output=%v", upgraded.Input, upgraded.Output)
	}
}

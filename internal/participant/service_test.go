package participant

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/store"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAvailableKeysExcludesEveryOtherParticipantAndKeepsCurrentFirst(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Close()
	service := Service{Store: db}
	current, err := service.Save(Input{Participant: domain.Participant{Name: "current"}, APIKey: "key-current"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := service.Save(Input{Participant: domain.Participant{Name: "other"}, APIKey: "key-used"})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Delete(other.ID); err != nil {
		t.Fatal(err)
	}

	got, err := service.AvailableKeys(
		[]string{" key-free ", "key-used", "key-current", "key-free", ""},
		current.ID,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"key-current", "key-free"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AvailableKeys() = %#v, want %#v", got, want)
	}
}

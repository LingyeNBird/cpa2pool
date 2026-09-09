package quota

import (
	"cpa2pool/internal/domain"
	"testing"
	"time"
)

func TestMonthlyBoundaryRestoresAnchorDayAfterFebruary(t *testing.T) {
	anchor := time.Date(2027, time.January, 31, 12, 30, 0, 0, time.UTC)
	q := domain.Quota{Period: "month", Anchor: anchor}
	feb := nextBoundary(q, anchor)
	march := nextBoundary(q, *feb)
	if feb.Day() != 28 || feb.Month() != time.February || march.Day() != 31 || march.Month() != time.March || march.Hour() != 12 {
		t.Fatalf("monthly boundaries drifted: %s -> %s", feb, march)
	}
}
func TestQuotaValidityUsesHalfOpenInterval(t *testing.T) {
	// Quota periods end at the exact boundary, not one request later.
	start := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	q := domain.Quota{Enabled: true, StartsAt: start, ExpiresAt: &end}
	if !Active(q, start) || Active(q, end) || Active(q, start.Add(-time.Nanosecond)) {
		t.Fatal("quota validity must be [start, end)")
	}
}

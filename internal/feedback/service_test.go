package feedback

import (
	"cpa2pool/internal/store"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func testService(t *testing.T, handler http.HandlerFunc) *Service {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.DB.Close() })
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return newService(db, srv.URL)
}

func TestSubmitDeliversImmediatelyWhenUpstreamAccepts(t *testing.T) {
	var calls int32
	s := testService(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		var body feedbackRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Project != "cpa2pool" || body.Content != "hello" {
			t.Errorf("unexpected request body: %+v", body)
		}
		if len(body.ClientID) < 16 || len(body.ClientID) > 32 {
			t.Errorf("clientId should be 16-32 chars, got %q", body.ClientID)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(feedbackResponse{Accepted: true})
	})
	if err := s.Submit("hello"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}
	if n := s.queueLen(); n != 0 {
		t.Fatalf("expected empty queue, got %d", n)
	}
}

func TestSubmitQueuesAndReportsSuccessWhenUpstreamRateLimits(t *testing.T) {
	s := testService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	if err := s.Submit("please retry me"); err != nil {
		t.Fatalf("Submit should report success even when upstream is rate limited, got %v", err)
	}
	if n := s.queueLen(); n != 1 {
		t.Fatalf("expected 1 queued item, got %d", n)
	}
}

func TestSubmitPreservesOrderWhileBacklogged(t *testing.T) {
	var calls int32
	s := testService(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	})
	if err := s.Submit("first"); err != nil {
		t.Fatal(err)
	}
	if err := s.Submit("second"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("second submission should not attempt immediate delivery while backlogged, got %d calls", calls)
	}
	if n := s.queueLen(); n != 2 {
		t.Fatalf("expected 2 queued items, got %d", n)
	}
}

func TestFlushDrainsOldestQueuedItemOnceUpstreamRecovers(t *testing.T) {
	var limited atomic.Bool
	limited.Store(true)
	s := testService(t, func(w http.ResponseWriter, r *http.Request) {
		if limited.Load() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(feedbackResponse{Accepted: true})
	})
	if err := s.Submit("queued"); err != nil {
		t.Fatal(err)
	}
	if n := s.queueLen(); n != 1 {
		t.Fatalf("expected 1 queued item, got %d", n)
	}
	s.flush()
	if n := s.queueLen(); n != 1 {
		t.Fatalf("flush should leave the item queued while still rate limited, got %d", n)
	}
	limited.Store(false)
	s.flush()
	if n := s.queueLen(); n != 0 {
		t.Fatalf("flush should deliver and dequeue once upstream recovers, got %d", n)
	}
}

func TestSubmitRejectsInvalidContent(t *testing.T) {
	s := testService(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("upstream should not be called for invalid input")
	})
	if err := s.Submit("   "); err == nil {
		t.Fatal("expected an error for empty content")
	}
	if err := s.Submit(strings.Repeat("x", MaxContentLen+1)); err == nil {
		t.Fatal("expected an error for content over the limit")
	}
}

func TestClientIDIsStablePerService(t *testing.T) {
	s := testService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(feedbackResponse{Accepted: true})
	})
	first := s.clientID()
	second := s.clientID()
	if first != second {
		t.Fatalf("clientID should be stable, got %q then %q", first, second)
	}
}

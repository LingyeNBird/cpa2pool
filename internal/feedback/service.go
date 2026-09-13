package feedback

import (
	"bytes"
	"cpa2pool/internal/domain"
	"cpa2pool/internal/store"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	endpoint      = "https://feedbackbash.nightunderfly.online/api/v1/feedback"
	project       = "cpa2pool"
	MaxContentLen = 800
	retryInterval = 5 * time.Minute
)

// Service submits user feedback to the upstream feedback endpoint. The
// upstream allows at most one submission per five minutes per source IP;
// callers should never see that as a failure. A submission that upstream
// rate-limits is queued and retried automatically by a background loop, so
// Submit only fails for genuinely invalid input or a request never worth
// retrying.
type Service struct {
	db       *store.Store
	client   *http.Client
	endpoint string
	stop     chan struct{}
}

func New(db *store.Store) *Service {
	s := newService(db, endpoint)
	go s.loop()
	return s
}

func newService(db *store.Store, url string) *Service {
	if _, err := db.DB.Exec(`
CREATE TABLE IF NOT EXISTS feedback_queue(id TEXT PRIMARY KEY,content TEXT NOT NULL,created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS feedback_meta(key TEXT PRIMARY KEY,value TEXT NOT NULL);`); err != nil {
		panic(err)
	}
	return &Service{db: db, client: &http.Client{Timeout: 10 * time.Second}, endpoint: url, stop: make(chan struct{})}
}

func (s *Service) Close() { close(s.stop) }

func (s *Service) loop() {
	s.flush()
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.flush()
		}
	}
}

// Submit delivers content immediately when nothing is backlogged. If the
// upstream needs to wait (rate limited, unreachable, ...) the content is
// queued for the background loop instead, and Submit still reports success:
// delivery is guaranteed eventually, just not necessarily right away.
func (s *Service) Submit(content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("反馈内容不能为空")
	}
	if len([]rune(content)) > MaxContentLen {
		return errors.New("反馈内容最多 800 字")
	}
	if s.queueLen() > 0 {
		// Something is already waiting its turn; preserve submission order
		// instead of letting a fresh item race ahead of the backlog.
		return s.enqueue(content)
	}
	if err := s.send(content); err != nil {
		return s.enqueue(content)
	}
	return nil
}

func (s *Service) queueLen() int {
	var n int
	_ = s.db.DB.QueryRow(`SELECT COUNT(*) FROM feedback_queue`).Scan(&n)
	return n
}

func (s *Service) enqueue(content string) error {
	_, err := s.db.DB.Exec(`INSERT INTO feedback_queue(id,content,created_at) VALUES(?,?,?)`,
		domain.ID(), content, domain.Now().Format("2006-01-02T15:04:05.000000000Z"))
	return err
}

// flush attempts to deliver the single oldest queued item. It leaves the
// queue untouched on failure so the next tick retries the same item first.
func (s *Service) flush() {
	var id, content string
	row := s.db.DB.QueryRow(`SELECT id,content FROM feedback_queue ORDER BY created_at LIMIT 1`)
	if err := row.Scan(&id, &content); err != nil {
		return
	}
	if err := s.send(content); err != nil {
		return
	}
	_, _ = s.db.DB.Exec(`DELETE FROM feedback_queue WHERE id=?`, id)
}

type feedbackRequest struct {
	Project  string `json:"project"`
	Content  string `json:"content"`
	ClientID string `json:"clientId"`
}
type feedbackResponse struct {
	Accepted bool `json:"accepted"`
}

// send makes one delivery attempt and returns nil only on a confirmed
// upstream accept.
func (s *Service) send(content string) error {
	body, err := json.Marshal(feedbackRequest{Project: project, Content: content, ClientID: s.clientID()})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("upstream status %d", resp.StatusCode)
	}
	var out feedbackResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || !out.Accepted {
		return errors.New("upstream did not accept feedback")
	}
	return nil
}

// clientID returns the persistent identifier sent with every submission so
// upstream can associate them with this deployment, generating and storing
// one on first use.
func (s *Service) clientID() string {
	var id string
	row := s.db.DB.QueryRow(`SELECT value FROM feedback_meta WHERE key='client_id'`)
	if err := row.Scan(&id); err == nil && id != "" {
		return id
	}
	id = domain.ID()
	_, _ = s.db.DB.Exec(`INSERT OR REPLACE INTO feedback_meta(key,value) VALUES('client_id',?)`, id)
	return id
}

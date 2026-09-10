package participant

import (
	"cpa2pool/internal/domain"
	"cpa2pool/internal/store"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
)

type Service struct{ Store *store.Store }
type Input struct {
	domain.Participant
	APIKey string `json:"api_key"`
}

func Scope(key string) string {
	sum := sha256.Sum256([]byte("cli-proxy-api:caller-scope:v1\x00" + strings.TrimSpace(key)))
	return hex.EncodeToString(sum[:])
}
func (s Service) List() ([]domain.Participant, error) {
	return store.List[domain.Participant](s.Store.DB, "SELECT body FROM participants WHERE json_extract(body,'$.deleted')=0 ORDER BY rowid DESC")
}
func (s Service) Get(id string) (domain.Participant, error) { return Get(s.Store.DB, id) }
func Get(q store.Query, id string) (domain.Participant, error) {
	return store.One[domain.Participant](q, "SELECT body FROM participants WHERE id=?", id)
}
func ByScope(q store.Query, scope string) (domain.Participant, error) {
	return store.One[domain.Participant](q, "SELECT body FROM participants WHERE scope=?", scope)
}
func (s Service) AvailableKeys(keys []string, currentID string) ([]string, error) {
	rows, err := s.Store.DB.Query("SELECT id,scope FROM participants")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	used := map[string]bool{}
	var currentScope string
	for rows.Next() {
		var id, scope string
		if err = rows.Scan(&id, &scope); err != nil {
			return nil, err
		}
		if id == currentID {
			currentScope = scope
		} else {
			used[scope] = true
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	available := make([]string, 0, len(keys))
	seen := map[string]bool{}
	var currentKey string
	for _, key := range keys {
		key = strings.TrimSpace(key)
		scope := Scope(key)
		if key == "" || seen[key] || used[scope] {
			continue
		}
		if scope == currentScope {
			currentKey = key
		} else {
			available = append(available, key)
		}
		seen[key] = true
	}
	if currentKey != "" {
		available = append([]string{currentKey}, available...)
	}
	return available, nil
}
func (s Service) Save(in Input) (domain.Participant, error) {
	var out domain.Participant
	err := s.Store.Tx(func(tx *sql.Tx) error {
		p := in.Participant
		var before any
		if strings.TrimSpace(p.Name) == "" {
			return errors.New("请输入参与者名称")
		}
		if p.ID == "" {
			p.ID = domain.ID()
			p.CreatedAt = domain.Now()
		} else {
			old, err := Get(tx, p.ID)
			if err != nil {
				return err
			}
			if old.Deleted {
				return errors.New("参与者已删除")
			}
			before = old
			p.CreatedAt = old.CreatedAt
			p.KeyScope = old.KeyScope
			p.KeyPreview = old.KeyPreview
		}
		key := strings.TrimSpace(in.APIKey)
		if key != "" {
			p.KeyScope = Scope(key)
			p.KeyPreview = "••••"
			if len(key) > 4 {
				p.KeyPreview += key[len(key)-4:]
			}
		}
		if p.KeyScope == "" {
			return errors.New("请关联 CPA API Key")
		}
		p.Deleted = false
		if p.Models == nil {
			p.Models = []string{}
		}
		if p.Efforts == nil {
			p.Efforts = []string{}
		}
		_, err := tx.Exec("INSERT INTO participants VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET scope=excluded.scope,body=excluded.body", p.ID, p.KeyScope, store.JSON(p))
		if err != nil {
			return err
		}
		out = p
		return store.Audit(tx, p.ID, "", "participant.save", "", before, p)
	})
	return out, err
}
func (s Service) Delete(id string) error {
	return s.Store.Tx(func(tx *sql.Tx) error {
		p, err := Get(tx, id)
		if err != nil {
			return err
		}
		before := p
		p.Deleted = true
		p.Enabled = false
		_, err = tx.Exec("UPDATE participants SET body=? WHERE id=?", store.JSON(p), id)
		if err != nil {
			return err
		}
		return store.Audit(tx, id, "", "participant.delete", "", before, p)
	})
}

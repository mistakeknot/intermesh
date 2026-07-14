package receipt

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mistakeknot/intermesh/internal/route"
)

type Writer struct {
	Path         string
	IncludeQuery bool
	Now          func() time.Time
}

type Metadata struct {
	CWD     string
	Limit   int
	Latency time.Duration
}

type Candidate struct {
	ID         string             `json:"id"`
	SkillMD    string             `json:"skill_md"`
	Score      float64            `json:"score"`
	Reasons    []string           `json:"reasons"`
	Components map[string]float64 `json:"components,omitempty"`
	SelectedBy string             `json:"selected_by"`
	RequiredBy []string           `json:"required_by"`
}

type Record struct {
	Event               string      `json:"event"`
	RouteID             string      `json:"route_id"`
	Timestamp           string      `json:"timestamp"`
	QueryHash           string      `json:"query_hash"`
	Query               string      `json:"query,omitempty"`
	Host                string      `json:"host,omitempty"`
	CWD                 string      `json:"cwd,omitempty"`
	RegistryFingerprint string      `json:"registry_fingerprint,omitempty"`
	Limit               int         `json:"limit"`
	Candidates          []Candidate `json:"candidates"`
	Warnings            []string    `json:"warnings"`
	LatencyMicros       int64       `json:"latency_micros"`
}

func (writer Writer) Append(ctx context.Context, result route.Result, metadata Metadata) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := writer.Path
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	identifier, err := randomID()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if writer.Now != nil {
		now = writer.Now().UTC()
	}
	normalizedQuery := strings.Join(strings.Fields(strings.ToLower(result.Query)), " ")
	hash := sha256.Sum256([]byte(normalizedQuery))
	candidates := make([]Candidate, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		candidates = append(candidates, Candidate{
			ID: candidate.ID, SkillMD: candidate.SkillMD, Score: candidate.Score,
			Reasons: candidate.Reasons, Components: candidate.Components,
			SelectedBy: candidate.SelectedBy, RequiredBy: candidate.RequiredBy,
		})
	}
	record := Record{
		Event: "intermesh.route.v1", RouteID: identifier, Timestamp: now.Format(time.RFC3339Nano),
		QueryHash: fmt.Sprintf("sha256:%x", hash), Host: result.Host, CWD: metadata.CWD,
		RegistryFingerprint: result.RegistryFingerprint, Limit: metadata.Limit,
		Candidates: candidates, Warnings: result.Warnings, LatencyMicros: metadata.Latency.Microseconds(),
	}
	if writer.IncludeQuery {
		record.Query = result.Query
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	_, err = file.Write(encoded)
	return err
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func DefaultPath() string {
	if stateHome := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); stateHome != "" {
		return filepath.Join(stateHome, "intermesh", "routes.jsonl")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".intermesh", "routes.jsonl")
	}
	return filepath.Join(home, ".local", "state", "intermesh", "routes.jsonl")
}

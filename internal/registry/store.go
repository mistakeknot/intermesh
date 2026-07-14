package registry

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mistakeknot/intermesh/internal/skill"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

type Store struct {
	db *sql.DB
}

type Generation struct {
	Fingerprint     string             `json:"fingerprint"`
	SkillCount      int                `json:"skill_count"`
	DiagnosticCount int                `json:"diagnostic_count"`
	Skills          []skill.Skill      `json:"skills,omitempty"`
	Diagnostics     []skill.Diagnostic `json:"diagnostics,omitempty"`
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("registry path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0o700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", absolute)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize registry: %w", err)
	}
	return &Store{db: db}, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) Replace(ctx context.Context, skills []skill.Skill, roots []skill.Root, diagnostics []skill.Diagnostic) (Generation, error) {
	normalizedSkills := append([]skill.Skill(nil), skills...)
	sort.Slice(normalizedSkills, func(i, j int) bool { return normalizedSkills[i].ID < normalizedSkills[j].ID })
	for index := 1; index < len(normalizedSkills); index++ {
		if normalizedSkills[index-1].ID == normalizedSkills[index].ID {
			return Generation{}, fmt.Errorf("duplicate skill ID %q", normalizedSkills[index].ID)
		}
	}
	normalizedRoots := append([]skill.Root(nil), roots...)
	sort.Slice(normalizedRoots, func(i, j int) bool {
		if normalizedRoots[i].Path == normalizedRoots[j].Path {
			return normalizedRoots[i].Namespace < normalizedRoots[j].Namespace
		}
		return normalizedRoots[i].Path < normalizedRoots[j].Path
	})
	normalizedDiagnostics := append([]skill.Diagnostic(nil), diagnostics...)
	sort.Slice(normalizedDiagnostics, func(i, j int) bool {
		left := normalizedDiagnostics[i]
		right := normalizedDiagnostics[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Message < right.Message
	})

	fingerprint, err := fingerprint(normalizedSkills, normalizedRoots)
	if err != nil {
		return Generation{}, err
	}

	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return Generation{}, err
	}
	defer func() { _ = transaction.Rollback() }()

	for _, table := range []string{"triggers", "edges", "diagnostics", "roots", "skills"} {
		if _, err := transaction.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return Generation{}, err
		}
	}
	if _, err := transaction.ExecContext(ctx, `INSERT INTO registry_meta(key, value) VALUES('fingerprint', ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, fingerprint); err != nil {
		return Generation{}, err
	}

	for _, root := range normalizedRoots {
		if _, err := transaction.ExecContext(ctx, `INSERT INTO roots(path, namespace) VALUES(?, ?)`, root.Path, root.Namespace); err != nil {
			return Generation{}, err
		}
	}
	for _, item := range normalizedSkills {
		manifestJSON, err := json.Marshal(item.Manifest)
		if err != nil {
			return Generation{}, err
		}
		if _, err := transaction.ExecContext(ctx, `INSERT INTO skills(id, namespace, name, description, skill_md, directory, body_hash, manifest_json) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.Namespace, item.Name, item.Description, item.SkillMD, item.Directory, item.BodyHash, string(manifestJSON)); err != nil {
			return Generation{}, err
		}
		if err := insertManifest(ctx, transaction, item); err != nil {
			return Generation{}, err
		}
	}
	for _, diagnostic := range normalizedDiagnostics {
		if _, err := transaction.ExecContext(ctx, `INSERT INTO diagnostics(path, code, message, severity) VALUES(?, ?, ?, ?)`, diagnostic.Path, diagnostic.Code, diagnostic.Message, diagnostic.Severity); err != nil {
			return Generation{}, err
		}
	}
	if err := transaction.Commit(); err != nil {
		return Generation{}, err
	}
	return Generation{Fingerprint: fingerprint, SkillCount: len(normalizedSkills), DiagnosticCount: len(normalizedDiagnostics)}, nil
}

func (store *Store) Current(ctx context.Context) (Generation, error) {
	var generation Generation
	if err := store.db.QueryRowContext(ctx, `SELECT value FROM registry_meta WHERE key='fingerprint'`).Scan(&generation.Fingerprint); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return generation, nil
		}
		return Generation{}, err
	}
	rows, err := store.db.QueryContext(ctx, `SELECT id, namespace, name, description, skill_md, directory, body_hash, manifest_json FROM skills ORDER BY id`)
	if err != nil {
		return Generation{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item skill.Skill
		var manifestJSON string
		if err := rows.Scan(&item.ID, &item.Namespace, &item.Name, &item.Description, &item.SkillMD, &item.Directory, &item.BodyHash, &manifestJSON); err != nil {
			return Generation{}, err
		}
		if err := json.Unmarshal([]byte(manifestJSON), &item.Manifest); err != nil {
			return Generation{}, err
		}
		generation.Skills = append(generation.Skills, item)
	}
	if err := rows.Err(); err != nil {
		return Generation{}, err
	}
	diagnosticRows, err := store.db.QueryContext(ctx, `SELECT path, code, message, severity FROM diagnostics ORDER BY path, code, message`)
	if err != nil {
		return Generation{}, err
	}
	defer diagnosticRows.Close()
	for diagnosticRows.Next() {
		var diagnostic skill.Diagnostic
		if err := diagnosticRows.Scan(&diagnostic.Path, &diagnostic.Code, &diagnostic.Message, &diagnostic.Severity); err != nil {
			return Generation{}, err
		}
		generation.Diagnostics = append(generation.Diagnostics, diagnostic)
	}
	generation.SkillCount = len(generation.Skills)
	generation.DiagnosticCount = len(generation.Diagnostics)
	return generation, diagnosticRows.Err()
}

func insertManifest(ctx context.Context, transaction *sql.Tx, item skill.Skill) error {
	triggers := map[string][]string{
		"phrase":      item.Manifest.Phrases,
		"extension":   item.Manifest.Extensions,
		"environment": item.Manifest.Environments,
	}
	for kind, values := range triggers {
		for _, value := range values {
			if _, err := transaction.ExecContext(ctx, `INSERT INTO triggers(skill_id, kind, value) VALUES(?, ?, ?)`, item.ID, kind, value); err != nil {
				return err
			}
		}
	}
	edges := map[string][]string{
		"requires":       item.Manifest.Requires,
		"composes_with":  item.Manifest.ComposesWith,
		"conflicts_with": item.Manifest.ConflictsWith,
		"supersedes":     item.Manifest.Supersedes,
	}
	for relation, targets := range edges {
		for _, target := range targets {
			if _, err := transaction.ExecContext(ctx, `INSERT INTO edges(source_id, relation, target_id) VALUES(?, ?, ?)`, item.ID, relation, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func fingerprint(skills []skill.Skill, roots []skill.Root) (string, error) {
	type record struct {
		Skills []skill.Skill `json:"skills"`
		Roots  []skill.Root  `json:"roots"`
	}
	encoded, err := json.Marshal(record{Skills: skills, Roots: roots})
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(encoded)
	return fmt.Sprintf("sha256:%x", hash), nil
}

func DefaultPath() string {
	if dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); dataHome != "" {
		return filepath.Join(dataHome, "intermesh", "registry.db")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".intermesh", "registry.db")
	}
	return filepath.Join(home, ".local", "share", "intermesh", "registry.db")
}

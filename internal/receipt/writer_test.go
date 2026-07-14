package receipt

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mistakeknot/intermesh/internal/route"
)

func TestAppendOmitsRawQueryByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "routes.jsonl")
	writer := Writer{Path: path, Now: func() time.Time { return time.Unix(10, 0).UTC() }}
	result := route.Result{
		Version: 1, Query: "secret user query", Host: "codex", RegistryFingerprint: "sha256:test",
		Candidates: []route.Candidate{{ID: "fixture:test", Description: "routing-only candidate metadata", ConflictsWith: []string{"fixture:private-neighbor"}, Score: 10}},
	}

	if err := writer.Append(context.Background(), result, Metadata{Latency: 3 * time.Millisecond, Limit: 5}); err != nil {
		t.Fatal(err)
	}
	receipt := readReceipt(t, path)
	if receipt.Event != "intermesh.route.v1" || receipt.QueryHash == "" {
		t.Fatalf("unexpected receipt: %#v", receipt)
	}
	if receipt.Query != "" {
		t.Fatalf("raw query leaked: %q", receipt.Query)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "routing-only candidate metadata") || strings.Contains(string(raw), `"description"`) {
		t.Fatalf("candidate description leaked into receipt: %s", raw)
	}
	if strings.Contains(string(raw), "fixture:private-neighbor") || strings.Contains(string(raw), `"conflicts_with"`) {
		t.Fatalf("candidate conflict metadata leaked into privacy-minimal receipt: %s", raw)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected receipt permissions: %o", info.Mode().Perm())
	}
}

func TestAppendCanExplicitlyRecordPlainQuery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "routes.jsonl")
	writer := Writer{Path: path, IncludeQuery: true}
	result := route.Result{Version: 1, Query: "plain query"}

	if err := writer.Append(context.Background(), result, Metadata{}); err != nil {
		t.Fatal(err)
	}
	if got := readReceipt(t, path).Query; got != "plain query" {
		t.Fatalf("expected explicit query, got %q", got)
	}
}

func readReceipt(t *testing.T, path string) Record {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatalf("missing receipt: %v", scanner.Err())
	}
	var record Record
	if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	return record
}

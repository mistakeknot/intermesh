package profile

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Apply(plan Plan, snapshotRoot string) (Snapshot, error) {
	if plan.Version != 1 {
		return Snapshot{}, fmt.Errorf("unsupported plan version %d", plan.Version)
	}
	if plan.Mode != ModeManaged {
		return Snapshot{}, errors.New("profile is routing-only; managed activation is unsafe")
	}
	if _, err := canonicalDirectory(plan.Router); err != nil {
		return Snapshot{}, fmt.Errorf("router: %w", err)
	}
	routerLink := filepath.Join(plan.Catalog, RouterEntryName)
	if target, err := os.Readlink(routerLink); err == nil {
		if target != plan.Router || !entriesAbsent(plan.Entries) {
			return Snapshot{}, fmt.Errorf("router entry already exists at %s with catalog drift", routerLink)
		}
		snapshot, err := snapshotForPlan(plan)
		if err != nil {
			return Snapshot{}, err
		}
		snapshotPath, err := writeSnapshot(snapshotRoot, snapshot)
		if err != nil {
			return Snapshot{}, err
		}
		snapshot.Path = snapshotPath
		return snapshot, nil
	} else if !os.IsNotExist(err) {
		return Snapshot{}, err
	}
	for _, entry := range plan.Entries {
		if !directChild(plan.Catalog, entry.Path) {
			return Snapshot{}, fmt.Errorf("entry %s escapes catalog %s", entry.Path, plan.Catalog)
		}
		target, err := os.Readlink(entry.Path)
		if err != nil {
			return Snapshot{}, fmt.Errorf("catalog drift at %s: %w", entry.Path, err)
		}
		if target != entry.Target {
			return Snapshot{}, fmt.Errorf("catalog drift at %s: target changed", entry.Path)
		}
	}
	snapshot, err := snapshotForPlan(plan)
	if err != nil {
		return Snapshot{}, err
	}
	snapshotPath, err := writeSnapshot(snapshotRoot, snapshot)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Path = snapshotPath

	removed := make([]Entry, 0, len(plan.Entries))
	rollback := func() {
		_ = os.Remove(routerLink)
		for _, entry := range removed {
			_ = os.Symlink(entry.Target, entry.Path)
		}
	}
	for _, entry := range plan.Entries {
		if err := os.Remove(entry.Path); err != nil {
			rollback()
			return Snapshot{}, err
		}
		removed = append(removed, entry)
	}
	if err := os.Symlink(plan.Router, routerLink); err != nil {
		rollback()
		return Snapshot{}, err
	}
	return snapshot, nil
}

func snapshotForPlan(plan Plan) (Snapshot, error) {
	snapshot := Snapshot{Version: 1, Host: plan.Host, Catalog: plan.Catalog, Router: plan.Router, Entries: append([]Entry(nil), plan.Entries...)}
	fingerprint, err := snapshotFingerprint(snapshot)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Fingerprint = fingerprint
	return snapshot, nil
}

func snapshotFingerprint(snapshot Snapshot) (string, error) {
	payload := struct {
		Version int     `json:"version"`
		Host    string  `json:"host"`
		Catalog string  `json:"catalog"`
		Router  string  `json:"router"`
		Entries []Entry `json:"entries"`
	}{snapshot.Version, snapshot.Host, snapshot.Catalog, snapshot.Router, snapshot.Entries}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func entriesAbsent(entries []Entry) bool {
	for _, entry := range entries {
		if _, err := os.Lstat(entry.Path); !os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func writeSnapshot(root string, snapshot Snapshot) (string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return "", err
	}
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	path := filepath.Join(absolute, "snapshot-"+hex.EncodeToString(random)+".json")
	encoded, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", err
	}
	encoded = append(encoded, '\n')
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, encoded, 0o600); err != nil {
		return "", err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return "", err
	}
	return path, nil
}

func directChild(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative == filepath.Base(child) && relative != "." && !filepath.IsAbs(relative)
}

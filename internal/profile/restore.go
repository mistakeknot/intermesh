package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Restore(snapshotPath string) error {
	content, err := os.ReadFile(snapshotPath)
	if err != nil {
		return err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(content, &snapshot); err != nil {
		return err
	}
	if snapshot.Version != 1 {
		return fmt.Errorf("unsupported snapshot version %d", snapshot.Version)
	}
	fingerprint, err := snapshotFingerprint(snapshot)
	if err != nil {
		return err
	}
	if snapshot.Fingerprint != fingerprint {
		return fmt.Errorf("snapshot integrity check failed")
	}
	routerLink := filepath.Join(snapshot.Catalog, RouterEntryName)
	target, err := os.Readlink(routerLink)
	if os.IsNotExist(err) && entriesMatch(snapshot.Entries) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("router catalog drift: %w", err)
	}
	if target != snapshot.Router {
		return fmt.Errorf("router catalog drift: target changed")
	}
	for _, entry := range snapshot.Entries {
		if !directChild(snapshot.Catalog, entry.Path) {
			return fmt.Errorf("entry %s escapes catalog", entry.Path)
		}
		if _, err := os.Lstat(entry.Path); err == nil {
			return fmt.Errorf("cannot restore over existing entry %s", entry.Path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Remove(routerLink); err != nil {
		return err
	}
	restored := make([]Entry, 0, len(snapshot.Entries))
	rollback := func() {
		for _, entry := range restored {
			_ = os.Remove(entry.Path)
		}
		_ = os.Symlink(snapshot.Router, routerLink)
	}
	for _, entry := range snapshot.Entries {
		if err := os.Symlink(entry.Target, entry.Path); err != nil {
			rollback()
			return err
		}
		restored = append(restored, entry)
	}
	return nil
}

func entriesMatch(entries []Entry) bool {
	for _, entry := range entries {
		target, err := os.Readlink(entry.Path)
		if err != nil || target != entry.Target {
			return false
		}
	}
	return true
}

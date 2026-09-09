package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestSaveConflict(t *testing.T) {
	file, err := NewFileStorage(filepath.Join(t.TempDir(), "urls.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.producer.Close()
	for name, s := range map[string]Storage{"memory": NewStorageRepo(), "file": file} {
		t.Run(name, func(t *testing.T) {
			if err := s.Save("first", "https://a.test"); err != nil {
				t.Fatal(err)
			}
			err := s.Save("second", "https://a.test")
			if !errors.Is(err, ErrConflict) {
				t.Fatalf("conflict: %v", err)
			}
			if _, err := s.Get("second"); !errors.Is(err, ErrKeyNotFound) {
				t.Fatal("duplicate saved")
			}
			if err := s.Save("first", "https://other.test"); !errors.Is(err, ErrShortURLExists) {
				t.Fatalf("key collision: %v", err)
			}
			keys, err := s.(BatchStorage).SaveBatch(context.Background(), []URLPair{{"batch", "https://a.test"}})
			if err != nil || len(keys) != 1 || keys[0] != "first" {
				t.Fatalf("batch: %v %v", keys, err)
			}
		})
	}
}

package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestDeleteOwnershipAndPersistence(t *testing.T) {
	for _, kind := range []string{"memory", "file"} {
		t.Run(kind, func(t *testing.T) {
			var s interface {
				Storage
				BatchStorage
				DeleteStorage
			} = NewStorageRepo()
			path := filepath.Join(t.TempDir(), "urls.json")
			if kind == "file" {
				f, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer f.producer.Close()
				s = f
			}
			ctx := context.Background()
			if err := s.Save("a", "https://a", "alice"); err != nil {
				t.Fatal(err)
			}
			if _, err := s.SaveBatch(ctx, []URLPair{{"b", "https://b"}}, "bob"); err != nil {
				t.Fatal(err)
			}
			if err := s.Save("other", "https://a", "bob"); !errors.Is(err, ErrConflict) {
				t.Fatal(err)
			}
			if err := s.DeleteBatch(ctx, []DeleteRequest{{"bob", "a"}, {"alice", "b"}, {"", "missing"}}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Get("a"); err != nil {
				t.Fatalf("noncreator deleted URL: %v", err)
			}
			if err := s.DeleteBatch(ctx, []DeleteRequest{{"alice", "a"}, {"alice", "a"}, {"alice", "missing"}}); err != nil {
				t.Fatal(err)
			}
			check := func(s Storage) {
				t.Helper()
				if _, err := s.Get("a"); !errors.Is(err, ErrDeleted) {
					t.Fatalf("deleted URL: %v", err)
				}
				if _, err := s.Get("b"); err != nil {
					t.Fatalf("foreign URL: %v", err)
				}
				if _, err := s.Get("missing"); !errors.Is(err, ErrKeyNotFound) {
					t.Fatalf("missing URL: %v", err)
				}
			}
			check(s)
			if kind == "file" {
				f, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer f.producer.Close()
				check(f)
			}
		})
	}
}

func TestFailedFileDeleteDoesNotPublish(t *testing.T) {
	s, err := NewFileStorage(filepath.Join(t.TempDir(), "urls.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save("a", "https://a", "alice"); err != nil {
		t.Fatal(err)
	}
	s.producer.Close()
	if err := s.DeleteBatch(context.Background(), []DeleteRequest{{"alice", "a"}}); err == nil {
		t.Fatal("expected write error")
	}
	if _, err := s.Get("a"); err != nil {
		t.Fatal(err)
	}
}

package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestUserURLSelection(t *testing.T) {
	for _, kind := range []string{"memory", "file"} {
		t.Run(kind, func(t *testing.T) {
			var s Storage = NewStorageRepo()
			path := filepath.Join(t.TempDir(), "urls.json")
			if kind == "file" {
				f, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer f.producer.Close()
				s = f
			}
			if err := s.Save("a", "https://a.test", "alice"); err != nil {
				t.Fatal(err)
			}
			if err := s.Save("b", "https://b.test", "bob"); err != nil {
				t.Fatal(err)
			}
			if err := s.Save("unused", "https://a.test", "bob"); !errors.Is(err, ErrConflict) {
				t.Fatal(err)
			}
			if kind == "file" {
				f, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer f.producer.Close()
				s = f
			}
			ctx := context.Background()
			urls, err := s.GetUserURLs(ctx, "alice")
			if err != nil || len(urls) != 1 || urls[0] != (URLPair{"a", "https://a.test"}) {
				t.Fatalf("alice: %v %v", urls, err)
			}
			urls[0].OriginalURL = "changed"
			again, err := s.GetUserURLs(ctx, "alice")
			if err != nil || again[0].OriginalURL != "https://a.test" {
				t.Fatal("result aliases storage")
			}
			urls, err = s.GetUserURLs(ctx, "bob")
			if err != nil || len(urls) != 2 {
				t.Fatalf("bob: %v %v", urls, err)
			}
			urls, err = s.GetUserURLs(ctx, "unknown")
			if err != nil || len(urls) != 0 {
				t.Fatalf("unknown: %v %v", urls, err)
			}
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			if _, err := s.GetUserURLs(cancelled, "alice"); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		})
	}
}

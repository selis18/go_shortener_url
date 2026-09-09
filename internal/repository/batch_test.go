package repository

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func TestBatchStorage(t *testing.T) {
	for _, kind := range []string{"memory", "file"} {
		t.Run(kind, func(t *testing.T) {
			var storage interface {
				Storage
				BatchStorage
			} = NewStorageRepo()
			path := filepath.Join(t.TempDir(), "urls.json")
			if kind == "file" {
				file, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer file.producer.Close()
				storage = file
			}
			ctx := context.Background()
			keys, err := storage.SaveBatch(ctx, []URLPair{{"a", "https://a.test"}, {"b", "https://b.test"}, {"c", "https://a.test"}})
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(keys) != "[a b a]" {
				t.Fatalf("keys: %v", keys)
			}
			_, err = storage.SaveBatch(ctx, []URLPair{{"d", "https://d.test"}, {"a", "https://other.test"}})
			if !errors.Is(err, ErrShortURLExists) {
				t.Fatalf("collision: %v", err)
			}
			if _, err := storage.Get("d"); !errors.Is(err, ErrKeyNotFound) {
				t.Fatal("partial batch saved")
			}
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			if _, err := storage.SaveBatch(cancelled, []URLPair{{"e", "https://e.test"}}); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					got, err := storage.SaveBatch(ctx, []URLPair{{fmt.Sprint(i), "https://shared.test"}})
					if err != nil {
						t.Error(err)
						return
					}
					if original, err := storage.Get(got[0]); err != nil || original != "https://shared.test" {
						t.Errorf("lookup: %q, %v", original, err)
					}
				}(i)
			}
			wg.Wait()
			if kind == "file" {
				reopened, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.producer.Close()
				if v, err := reopened.Get("a"); err != nil || v != "https://a.test" {
					t.Fatalf("reload: %q %v", v, err)
				}
				if _, err := reopened.Get("d"); !errors.Is(err, ErrKeyNotFound) {
					t.Fatal("failed batch persisted")
				}
			}
		})
	}
}

func TestFileBatchWriteFailure(t *testing.T) {
	s, err := NewFileStorage(filepath.Join(t.TempDir(), "urls.json"))
	if err != nil {
		t.Fatal(err)
	}
	s.producer.Close()
	if _, err := s.SaveBatch(context.Background(), []URLPair{{"a", "https://a.test"}}); err == nil {
		t.Fatal("expected write error")
	}
	if _, err := s.Get("a"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatal("failed write published in memory")
	}
}

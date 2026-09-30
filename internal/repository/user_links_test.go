package repository

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestUserLinks(t *testing.T) {
	for _, kind := range []string{"memory", "file"} {
		t.Run(kind, func(t *testing.T) {
			memory := NewStorageRepo()
			var s interface {
				Storage
				BatchStorage
			} = memory
			path := filepath.Join(t.TempDir(), "urls.json")
			if kind == "file" {
				file, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer file.producer.Close()
				s, memory = file, file.storage
			}
			if err := s.Save("a", "https://a.test", "alice"); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				if err := s.Save("unused", "https://a.test", "bob"); !errors.Is(err, ErrConflict) {
					t.Fatal(err)
				}
			}
			keys, err := s.SaveBatch(context.Background(), []URLPair{{"unused", "https://a.test"}, {"b", "https://b.test"}, {"again", "https://b.test"}}, "carol")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(keys, []string{"a", "b", "b"}) {
				t.Fatal(keys)
			}
			if _, err := s.SaveBatch(context.Background(), []URLPair{{"unused", "https://a.test"}, {"b", "https://collision.test"}}, "failed"); !errors.Is(err, ErrShortURLExists) {
				t.Fatal(err)
			}
			want := map[string][]string{"alice": {"a"}, "bob": {"a"}, "carol": {"a", "b"}}
			if !reflect.DeepEqual(memory.uStorage, want) {
				t.Fatalf("links: %v", memory.uStorage)
			}
			if kind == "file" {
				reopened, err := NewFileStorage(path)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.producer.Close()
				for user, links := range want {
					got := reopened.storage.uStorage[user]
					if len(got) != len(links) {
						t.Fatalf("reload %s: %v", user, got)
					}
					for _, key := range links {
						found := false
						for _, actual := range got {
							if actual == key {
								found = true
							}
						}
						if !found {
							t.Fatalf("reload %s missing %s", user, key)
						}
					}
				}
				if len(reopened.storage.uStorage["failed"]) != 0 {
					t.Fatal("failed batch persisted ownership")
				}
			}
		})
	}
}

func TestUserLinkWriteFailure(t *testing.T) {
	s, err := NewFileStorage(filepath.Join(t.TempDir(), "urls.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save("a", "https://a.test", "alice"); err != nil {
		t.Fatal(err)
	}
	s.producer.Close()
	if err := s.Save("b", "https://a.test", "bob"); err == nil || errors.Is(err, ErrConflict) {
		t.Fatalf("expected write error: %v", err)
	}
	if _, err := s.SaveBatch(context.Background(), []URLPair{{"b", "https://a.test"}}, "carol"); err == nil {
		t.Fatal("expected write error")
	}
	if len(s.storage.uStorage["bob"]) != 0 || len(s.storage.uStorage["carol"]) != 0 {
		t.Fatal("failed write published ownership")
	}
}

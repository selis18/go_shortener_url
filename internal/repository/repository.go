package repository

import (
	"context"
	"errors"
	"sync"
)

var ErrShortURLExists = errors.New("this url already is done")
var ErrShortURLEmpty = errors.New("url is empty")
var ErrKeyNotFound = errors.New("the key is not found")

var ErrConflict = errors.New("original URL already exists")

type Storage interface {
	Save(k string, v string) error
	Get(k string) (string, error)
	FindByValue(v string) (string, bool)
}
type ContextStorage interface {
	SaveContext(context.Context, string, string) error
	GetContext(context.Context, string) (string, error)
	FindByValueContext(context.Context, string) (string, bool, error)
}
type StorageRepo struct {
	mutex   sync.Mutex
	storage map[string]string
}

func NewStorageRepo() *StorageRepo {
	return &StorageRepo{
		storage: make(map[string]string),
	}
}
func (s *StorageRepo) Save(k string, v string) error {
	return s.SaveContext(context.Background(), k, v)
}
func (s *StorageRepo) SaveContext(ctx context.Context, k string, v string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v == "" {
		return ErrShortURLEmpty
	}
	for _, value := range s.storage {
		if value == v {
			return ErrConflict
		}
	}

	if _, e := s.storage[k]; e {
		return ErrShortURLExists
	}

	s.storage[k] = v
	return nil
}

func (s *StorageRepo) Get(k string) (string, error) {
	return s.GetContext(context.Background(), k)
}
func (s *StorageRepo) GetContext(ctx context.Context, k string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, e := s.storage[k]; e {
		return v, nil
	}
	return "", ErrKeyNotFound
}

func (s *StorageRepo) FindByValue(v string) (string, bool) {
	k, found, _ := s.FindByValueContext(context.Background(), v)
	return k, found
}
func (s *StorageRepo) FindByValueContext(ctx context.Context, v string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, value := range s.storage {
		if value == v {
			return key, true, nil
		}
	}
	return "", false, nil
}

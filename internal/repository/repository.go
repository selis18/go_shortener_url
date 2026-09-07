package repository

import (
	"errors"
	"sync"
)

var ErrShortURLExists = errors.New("this url already is done")
var ErrShortURLEmpty = errors.New("url is empty")
var ErrKeyNotFound = errors.New("the key is not found")

type Storage interface {
	Save(k string, v string) error
	Get(k string) (string, error)
	FindByValue(v string) (string, bool)
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v == "" {
		return ErrShortURLEmpty
	}

	if _, e := s.storage[k]; e {
		return ErrShortURLExists
	}

	s.storage[k] = v
	return nil
}

func (s *StorageRepo) Get(k string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, e := s.storage[k]; e {
		return v, nil
	}
	return "", ErrKeyNotFound
}

func (s *StorageRepo) FindByValue(v string) (string, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, value := range s.storage {
		if value == v {
			return key, true
		}
	}
	return "", false
}

package repository

import "fmt"

type Storage interface {
	Save(k string, v string) error
	Get(k string) (string, error)
}
type StorageRepo struct {
	storage map[string]string
}

func NewStorageRepo() *StorageRepo {
	return &StorageRepo{
		storage: make(map[string]string),
	}
}
func (s *StorageRepo) Save(k string, v string) error {
	if v == "" {
		return fmt.Errorf("Ссылка не может быть пустой!")
	}
	if _, exist := s.storage[v]; exist {
		return fmt.Errorf("Такая ссылка уже есть!")
	}

	if _, e := s.Get(k); e != nil {
		return fmt.Errorf("Такая короткая ссылка уже есть!")
	}
	s.storage[k] = v
	return nil
}

func (s *StorageRepo) Get(k string) (string, error) {
	if v, e := s.storage[k]; e {
		return v, nil
	}
	return "", fmt.Errorf("Ошибка на уровне поиска!")
}

package main

import "fmt"

type Storage interface {
	Save(k string, v string)
	Get(k string) error
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

	if _, e := s.storage[k]; e {
		return s.Save(k, v)
	}
	s.storage[k] = v
	return nil
}

func (s *StorageRepo) Get(k string) (string, error) {
	for key, value := range s.storage {
		if key == k {
			return value, nil
		} else {
			return "", fmt.Errorf("Не найдено!")
		}
	}
	return "", fmt.Errorf("Ошибка на уровне поиска!")
}

var StorageR = NewStorageRepo()

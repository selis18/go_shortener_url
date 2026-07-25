package main

import "fmt"

type Storage interface {
	Save(k string, v userUrl)
	Get(k string) error
}
type StorageRepo struct {
	storage map[string]userUrl
}

func NewStorageRepo() *StorageRepo {
	return &StorageRepo{
		storage: make(map[string]userUrl),
	}
}
func (s *StorageRepo) Save(k string, v userUrl) error {
	if v.inputUrl == "" {
		return fmt.Errorf("Ссылка не может быть пустой!")
	}
	if _, exist := s.storage[v.inputUrl]; exist {
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
			return value.inputUrl, nil
		} else {
			return "", fmt.Errorf("Не найдено!")
		}
	}
	return "", fmt.Errorf("Ошибка на уровне поиска!")
}

var StorageR = NewStorageRepo()

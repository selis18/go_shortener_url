package repository

import "fmt"

type Storage interface {
	Save(k string, v string) error
	Get(k string) (string, error)
}
type StorageRepo struct {
	Storage map[string]string
}

func NewStorageRepo() *StorageRepo {
	return &StorageRepo{
		Storage: make(map[string]string),
	}
}
func (s *StorageRepo) Save(k string, v string) error {
	if v == "" {
		return fmt.Errorf("Ссылка не может быть пустой!")
	}
	if _, exist := s.Storage[v]; exist {
		return fmt.Errorf("Такая ссылка уже есть!")
	}

	if _, e := s.Storage[k]; e {
		return s.Save(k, v)
	}
	s.Storage[k] = v
	return nil
}

func (s *StorageRepo) Get(k string) (string, error) {
	if v, e := s.Storage[k]; e {
		return v, nil
	}
	return "", fmt.Errorf("Ошибка на уровне поиска!")
}

var StorageR = NewStorageRepo()

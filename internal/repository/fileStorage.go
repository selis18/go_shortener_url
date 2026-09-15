package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/selis18/go_shortener_url/internal/model"
)

type Producer struct {
	file *os.File
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file: file,
	}, nil
}

func (p *Producer) WriteURL(URL *model.JSONStorage) error {
	data, err := json.Marshal(&URL)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = p.file.Write(data)
	return err
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadURL() (*model.JSONStorage, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}

	data := c.scanner.Bytes()

	jsonStorage := model.JSONStorage{}
	if err := json.Unmarshal(data, &jsonStorage); err != nil {
		return nil, err
	}

	return &jsonStorage, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

type FileStorage struct {
	mutex    sync.Mutex
	storage  *StorageRepo
	producer *Producer
	nextUUID int
}

func NewFileStorage(fileName string) (*FileStorage, error) {
	storage := FileStorage{
		storage:  NewStorageRepo(),
		nextUUID: 1,
	}

	consumer, err := NewConsumer(fileName)
	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	for {
		item, err := consumer.ReadURL()
		if err != nil {
			return nil, err
		}

		if item == nil {
			break
		}

		if item.OriginalURL == "" {
			return nil, ErrShortURLEmpty
		}
		if _, exists := storage.storage.storage[item.ShortURL]; !exists {
			storage.storage.storage[item.ShortURL] = item.OriginalURL
		}

		storage.storage.linkUser(item.UserID, item.ShortURL)
		id, err := strconv.Atoi(item.UUID)
		if err != nil {
			return nil, err
		}
		if id >= storage.nextUUID {
			storage.nextUUID = id + 1
		}
	}
	storage.producer, err = NewProducer(fileName)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func (s *FileStorage) Save(shortURL string, originalURL string, userID string) error {
	return s.SaveContext(context.Background(), shortURL, originalURL, userID)
}
func (s *FileStorage) SaveContext(ctx context.Context, shortURL string, originalURL string, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.storage.mutex.Lock()
	defer s.storage.mutex.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if originalURL == "" {
		return ErrShortURLEmpty
	}
	conflict := false
	for key, value := range s.storage.storage {
		if value == originalURL {
			shortURL = key
			conflict = true
			break
		}
	}
	if !conflict {
		if _, exists := s.storage.storage[shortURL]; exists {
			return ErrShortURLExists
		}
	} else {
		if userID == "" {
			return ErrConflict
		}
		for _, key := range s.storage.uStorage[userID] {
			if key == shortURL {
				return ErrConflict
			}
		}
	}

	u := &model.JSONStorage{
		UserID:      userID,
		UUID:        strconv.Itoa(s.nextUUID),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	if err := s.appendBatch(append(data, '\n')); err != nil {
		return err
	}

	s.storage.storage[shortURL] = originalURL
	s.storage.linkUser(userID, shortURL)
	s.nextUUID++
	if conflict {
		return ErrConflict
	}
	return nil
}

func (s *FileStorage) appendBatch(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	info, err := s.producer.file.Stat()
	if err != nil {
		return err
	}
	n, err := s.producer.file.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return errors.Join(err, s.producer.file.Truncate(info.Size()))
	}
	return nil
}

func (s *FileStorage) SaveBatch(ctx context.Context, pairs []URLPair, userID string) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.storage.mutex.Lock()
	defer s.storage.mutex.Unlock()
	keys, added, err := prepareBatch(ctx, s.storage.storage, pairs)
	if err != nil {
		return nil, err
	}
	// Persist both new URLs and new ownership of existing URLs.
	records := append([]URLPair(nil), added...)
	seen := make(map[string]bool)
	for _, pair := range added {
		seen[pair.ShortURL] = true
	}
	if userID != "" {
		for _, key := range s.storage.uStorage[userID] {
			seen[key] = true
		}
		for i, key := range keys {
			if !seen[key] {
				records = append(records, URLPair{ShortURL: key, OriginalURL: pairs[i].OriginalURL})
				seen[key] = true
			}
		}
	}
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	for i, pair := range records {
		if err := encoder.Encode(model.JSONStorage{
			UserID: userID, UUID: strconv.Itoa(s.nextUUID + i), ShortURL: pair.ShortURL, OriginalURL: pair.OriginalURL,
		}); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.appendBatch(data.Bytes()); err != nil {
		return nil, err
	}
	for _, pair := range added {
		s.storage.storage[pair.ShortURL] = pair.OriginalURL
	}
	for _, key := range keys {
		s.storage.linkUser(userID, key)
	}
	s.nextUUID += len(records)
	return keys, nil
}

func (s *FileStorage) Get(shortURL string) (string, error) {
	return s.GetContext(context.Background(), shortURL)
}
func (s *FileStorage) GetContext(ctx context.Context, shortURL string) (string, error) {
	return s.storage.GetContext(ctx, shortURL)
}

func (s *FileStorage) FindByValue(originalURL string) (string, bool) {
	k, found, _ := s.FindByValueContext(context.Background(), originalURL)
	return k, found
}
func (s *FileStorage) FindByValueContext(ctx context.Context, originalURL string) (string, bool, error) {
	return s.storage.FindByValueContext(ctx, originalURL)
}

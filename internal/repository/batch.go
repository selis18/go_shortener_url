package repository

import "context"

type URLPair struct {
	ShortURL    string
	OriginalURL string
}

type BatchStorage interface {
	SaveBatch(context.Context, []URLPair) ([]string, error)
}

func prepareBatch(ctx context.Context, current map[string]string, pairs []URLPair) ([]string, []URLPair, error) {
	byValue := make(map[string]string, len(current))
	for k, v := range current {
		byValue[v] = k
	}
	keys := make([]string, len(pairs))
	added := make([]URLPair, 0, len(pairs))
	reserved := make(map[string]bool)
	for i, pair := range pairs {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if pair.OriginalURL == "" {
			return nil, nil, ErrShortURLEmpty
		}
		if key, ok := byValue[pair.OriginalURL]; ok {
			keys[i] = key
			continue
		}
		if _, ok := current[pair.ShortURL]; ok {
			return nil, nil, ErrShortURLExists
		}
		if reserved[pair.ShortURL] {
			return nil, nil, ErrShortURLExists
		}
		reserved[pair.ShortURL] = true
		byValue[pair.OriginalURL] = pair.ShortURL
		keys[i] = pair.ShortURL
		added = append(added, pair)
	}
	return keys, added, ctx.Err()
}

func (s *StorageRepo) SaveBatch(ctx context.Context, pairs []URLPair) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	keys, added, err := prepareBatch(ctx, s.storage, pairs)
	if err != nil {
		return nil, err
	}
	for _, pair := range added {
		s.storage[pair.ShortURL] = pair.OriginalURL
	}
	return keys, nil
}

package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type Storage struct {
	mu   sync.Mutex
	path string
}

func New(path string) *Storage {
	return &Storage{
		path: path,
	}
}

// AppendWithKey appends value wrapped in a JSON object under key.
func (s *Storage) AppendWithKey(key string, value any) error {
	return s.append(key, value)
}

// append writes value as one JSONL record, optionally wrapped by key.
func (s *Storage) append(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(
		s.path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if key != "" {
		value = map[string]any{
			key: value,
		}
	}

	return json.NewEncoder(file).Encode(value)
}

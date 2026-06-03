package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendWithKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.jsonl")
	storage := New(path)

	err := storage.AppendWithKey("page 1", map[string]any{
		"id":   "123",
		"name": "Test Project",
	})

	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"page 1":{"id":"123","name":"Test Project"}}`,
		string(content),
	)
}

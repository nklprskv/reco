package poller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reco/pkg/client"
	"reco/pkg/storage"
)

func TestPollProjects(t *testing.T) {
	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("authorization"))
		assert.Equal(t, "test-workspace", r.URL.Query().Get("workspace"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))

		requests = append(requests, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("offset") {
		case "":
			fmt.Fprint(w, `{
				"data": [{"gid": "project-1", "name": "Project 1"}],
				"next_page": {
					"offset": "next-offset",
					"path": "/api/1.0/projects?offset=next-offset",
					"uri": "https://app.asana.com/api/1.0/projects?offset=next-offset"
				}
			}`)
		case "next-offset":
			fmt.Fprint(w, `{
				"data": [{"gid": "project-2", "name": "Project 2"}]
			}`)
		default:
			t.Fatalf("unexpected offset: %s", r.URL.Query().Get("offset"))
		}
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "projects.jsonl")
	projectsStorage := storage.New(path)
	usersStorage := storage.New(filepath.Join(t.TempDir(), "users.jsonl"))
	asanaClient := client.New(server.URL, "test-token")
	asanaPoller := New(
		asanaClient,
		"test-workspace",
		projectsStorage,
		usersStorage,
		time.Nanosecond,
		1,
		1,
	)

	err := asanaPoller.PollProjects()
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	assert.Len(t, requests, 2)
	require.Len(t, lines, 2)
	assert.Contains(t, requests[0], "workspace=test-workspace")
	assert.Contains(t, requests[0], "limit=1")
	assert.Contains(t, requests[1], "offset=next-offset")
	assert.Contains(t, requests[1], "workspace=test-workspace")
	assert.Contains(t, requests[1], "limit=1")
	assert.JSONEq(
		t,
		`{"page 1":{"data":[{"gid":"project-1","name":"Project 1"}],"next_page":{"offset":"next-offset","path":"/api/1.0/projects?offset=next-offset","uri":"https://app.asana.com/api/1.0/projects?offset=next-offset"}}}`,
		lines[0],
	)
	assert.JSONEq(
		t,
		`{"page 2":{"data":[{"gid":"project-2","name":"Project 2"}]}}`,
		lines[1],
	)
}

func TestPollUsers(t *testing.T) {
	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("authorization"))
		assert.Equal(t, "test-workspace", r.URL.Query().Get("workspace"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))

		requests = append(requests, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("offset") {
		case "":
			fmt.Fprint(w, `{
				"data": [{"gid": "user-1", "name": "User 1"}],
				"next_page": {
					"offset": "next-offset",
					"path": "/api/1.0/users?offset=next-offset",
					"uri": "https://app.asana.com/api/1.0/users?offset=next-offset"
				}
			}`)
		case "next-offset":
			fmt.Fprint(w, `{
				"data": [{"gid": "user-2", "name": "User 2"}]
			}`)
		default:
			t.Fatalf("unexpected offset: %s", r.URL.Query().Get("offset"))
		}
	}))
	defer server.Close()

	projectsStorage := storage.New(filepath.Join(t.TempDir(), "projects.jsonl"))
	path := filepath.Join(t.TempDir(), "users.jsonl")
	usersStorage := storage.New(path)
	asanaClient := client.New(server.URL, "test-token")
	asanaPoller := New(
		asanaClient,
		"test-workspace",
		projectsStorage,
		usersStorage,
		time.Nanosecond,
		1,
		1,
	)

	err := asanaPoller.PollUsers()
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	assert.Len(t, requests, 2)
	require.Len(t, lines, 2)
	assert.Contains(t, requests[0], "workspace=test-workspace")
	assert.Contains(t, requests[0], "limit=1")
	assert.Contains(t, requests[1], "offset=next-offset")
	assert.Contains(t, requests[1], "workspace=test-workspace")
	assert.Contains(t, requests[1], "limit=1")
	assert.JSONEq(
		t,
		`{"page 1":{"data":[{"gid":"user-1","name":"User 1"}],"next_page":{"offset":"next-offset","path":"/api/1.0/users?offset=next-offset","uri":"https://app.asana.com/api/1.0/users?offset=next-offset"}}}`,
		lines[0],
	)
	assert.JSONEq(
		t,
		`{"page 2":{"data":[{"gid":"user-2","name":"User 2"}]}}`,
		lines[1],
	)
}

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjects(t *testing.T) {
	var request *http.Request

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request = r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "test-token")
	response, err := client.GetProjects("test-workspace", 50, "test-offset")

	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	require.NotNil(t, request)
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "/projects", request.URL.Path)
	assert.Equal(t, "application/json", request.Header.Get("accept"))
	assert.Equal(t, "Bearer test-token", request.Header.Get("authorization"))
	assert.Equal(t, "test-workspace", request.URL.Query().Get("workspace"))
	assert.Equal(t, "50", request.URL.Query().Get("limit"))
	assert.Equal(t, "test-offset", request.URL.Query().Get("offset"))
}

func TestGetUsers(t *testing.T) {
	var request *http.Request

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request = r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "test-token")
	response, err := client.GetUsers("test-workspace", 25, "test-offset")

	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	require.NotNil(t, request)
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "/users", request.URL.Path)
	assert.Equal(t, "application/json", request.Header.Get("accept"))
	assert.Equal(t, "Bearer test-token", request.Header.Get("authorization"))
	assert.Equal(t, "test-workspace", request.URL.Query().Get("workspace"))
	assert.Equal(t, "25", request.URL.Query().Get("limit"))
	assert.Equal(t, "test-offset", request.URL.Query().Get("offset"))
}

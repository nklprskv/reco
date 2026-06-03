package client

import (
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	BaseUrl string
	Token   string
}

// New creates an Asana API client.
func New(baseUrl, token string) *Client {
	return &Client{
		BaseUrl: baseUrl,
		Token:   token,
	}
}

// GetProjects requests a paginated projects page.
func (c *Client) GetProjects(workspace string, limit int, offset string) (*http.Response, error) {
	url := strings.TrimRight(c.BaseUrl, "/") + "/projects"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	setPaginationQuery(req, workspace, limit, offset)
	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", "Bearer "+c.Token)

	return http.DefaultClient.Do(req)
}

// GetUsers requests a paginated users page.
func (c *Client) GetUsers(workspace string, limit int, offset string) (*http.Response, error) {
	url := strings.TrimRight(c.BaseUrl, "/") + "/users"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	setPaginationQuery(req, workspace, limit, offset)
	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", "Bearer "+c.Token)

	return http.DefaultClient.Do(req)
}

// setPaginationQuery applies workspace and pagination query params.
func setPaginationQuery(req *http.Request, workspace string, limit int, offset string) {
	query := req.URL.Query()

	if workspace != "" {
		query.Set("workspace", workspace)
	}

	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	if offset != "" {
		query.Set("offset", offset)
	}

	req.URL.RawQuery = query.Encode()
}

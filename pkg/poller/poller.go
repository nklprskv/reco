package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"reco/pkg/client"
	"reco/pkg/storage"
)

type Poller struct {
	client          *client.Client
	workspace       string
	projectsStorage *storage.Storage
	usersStorage    *storage.Storage
	every           time.Duration
	limiter         *rate.Limiter
	attempts        int
	pageLimit       int
}

type apiPage struct {
	Data     any       `json:"data"`
	NextPage *nextPage `json:"next_page,omitempty"`
}

type nextPage struct {
	Offset string `json:"offset"`
	Path   string `json:"path"`
	URI    string `json:"uri"`
}

// New creates a poller with rate limiting, retry, and pagination settings.
func New(
	client *client.Client,
	workspace string,
	projectsStorage *storage.Storage,
	usersStorage *storage.Storage,
	every time.Duration,
	attempts int,
	pageLimit int,
) *Poller {
	return &Poller{
		client:          client,
		workspace:       workspace,
		projectsStorage: projectsStorage,
		usersStorage:    usersStorage,
		every:           every,
		limiter:         rate.NewLimiter(rate.Every(every), 1),
		attempts:        attempts,
		pageLimit:       pageLimit,
	}
}

// PollProjects fetches and stores all paginated project pages.
func (p *Poller) PollProjects() error {
	return p.pollPaginated(
		"get projects",
		p.projectsStorage,
		p.client.GetProjects,
	)
}

// PollUsers fetches and stores all paginated user pages.
func (p *Poller) PollUsers() error {
	return p.pollPaginated(
		"get users",
		p.usersStorage,
		p.client.GetUsers,
	)
}

// pollPaginated follows next_page offsets and stores each page.
func (p *Poller) pollPaginated(
	action string,
	storage *storage.Storage,
	request func(workspace string, limit int, offset string) (*http.Response, error),
) error {
	pageNumber := 1
	offset := ""

	for {
		page, err := p.doRequest(action, func() (*http.Response, error) {
			return request(p.workspace, p.pageLimit, offset)
		})
		if err != nil {
			return err
		}

		if err := storage.AppendWithKey(
			fmt.Sprintf("page %d", pageNumber),
			page,
		); err != nil {
			return err
		}

		slog.Info("page appended", "action", action, "page", pageNumber)

		if page.NextPage == nil || page.NextPage.Offset == "" {
			return nil
		}

		offset = page.NextPage.Offset
		pageNumber++
	}
}

// doRequest runs one request with rate limiting and retry handling.
func (p *Poller) doRequest(
	action string,
	request func() (*http.Response, error),
) (*apiPage, error) {
	var lastErr error

	for attempt := 0; attempt < p.attempts; attempt++ {
		if err := p.limiter.Wait(context.Background()); err != nil {
			return nil, err
		}

		response, err := request()
		if err != nil {
			lastErr = err
			slog.Warn(
				"request failed, retrying",
				"action", action,
				"attempt", attempt+1,
				"attempts", p.attempts,
				"error", err,
			)
			continue
		}

		if response.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(response.Header.Get("Retry-After"))
			response.Body.Close()

			if attempt == p.attempts-1 {
				return nil, fmt.Errorf("%s failed: too many requests", action)
			}

			delay := retryAfter * time.Duration(1<<attempt)
			slog.Warn(
				"rate limited, retrying",
				"action", action,
				"attempt", attempt+1,
				"attempts", p.attempts,
				"delay", delay.String(),
			)
			time.Sleep(delay)
			continue
		}

		return decodeResponse(response, action)
	}

	return nil, lastErr
}

// decodeResponse decodes a successful Asana page or returns an API error.
func decodeResponse(response *http.Response, action string) (*apiPage, error) {
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %s", action, response.Status)
		}

		return nil, fmt.Errorf("%s failed: %s: %s", action, response.Status, string(body))
	}

	var value apiPage
	if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
		return nil, err
	}

	return &value, nil
}

// parseRetryAfter converts Retry-After into a delay.
func parseRetryAfter(value string) time.Duration {
	seconds, err := strconv.Atoi(value)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}

	retryAt, err := http.ParseTime(value)
	if err == nil {
		return time.Until(retryAt)
	}

	return time.Second
}

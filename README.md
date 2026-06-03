# Asana Extractor

This project extracts data from Asana and writes API responses into local JSONL files.

## Configuration

The app is configured via environment variables:

- `ASANA_BASE_URL` - required. Base Asana API URL, for example `https://app.asana.com/api/1.0`.
- `ASANA_TOKEN` - required. Personal access token used as the bearer token.
- `ASANA_WORKSPACE` - required. Workspace id used for paginated project and user requests.
- `FREQUENCY` - required. Request frequency in seconds.
- `ATTEMPTS` - optional. Retry attempts for failed/rate-limited requests. Defaults to `3`.
- `PAGE_LIMIT` - optional. Asana page size from `1` to `100`. Defaults to `100`.

## Run And Test

Run the extractor:

```sh
make run
```

Run tests:

```sh
make test
```

## Main Approach

The app initializes an Asana client, two JSONL storages, and a poller. The poller fetches projects and users from Asana, follows pagination when needed, and appends each page to the matching output file:

- `projects.jsonl`
- `users.jsonl`

Each storage owns its file path and mutex, so appends are serialized per file.

## Rate Limiting

The poller uses `golang.org/x/time/rate` with:

```go
rate.NewLimiter(rate.Every(every), 1)
```

`every` is built from `FREQUENCY`, so each request waits for the configured interval before calling Asana.

For `429 Too Many Requests`, the app reads Asana's `Retry-After` header and retries with exponential backoff. The number of attempts is controlled by `ATTEMPTS`.

## Pagination

Requests send:

- `workspace`
- `limit`
- `offset` when a next page exists

If Asana returns `next_page.offset`, the poller requests the next page. Each page is written as a separate JSONL line with the page number as a JSON key. This keeps memory usage low because the app stores one page at a time instead of keeping the full dataset in memory:

```json
{"page 1": {...}}
{"page 2": {...}}
```

# Athena

[![CI](https://github.com/akhil-datla/Athena/actions/workflows/ci.yml/badge.svg)](https://github.com/akhil-datla/Athena/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/akhil-datla/athena)](https://goreportcard.com/report/github.com/akhil-datla/athena)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Lightning-fast Q&A search engine.** Load your FAQ data from CSV or JSON, then search instantly via an interactive TUI, REST API, or beautiful web UI. Results are ranked by fuzzy matching with relevance scoring.

```
   ___   __  __
  / _ | / /_/ /  ___ ___  ___ _
 / __ |/ __/ _ \/ -_) _ \/ _ '/
/_/ |_|\__/_//_/\__/_//_/\_,_/
```

## Features

- **Interactive TUI** — live fuzzy search in your terminal (like fzf for Q&A)
- **Web UI** — beautiful, responsive search interface served at `/`
- **REST API** — JSON responses with CORS, pagination, health checks
- **Hot reload** — `--watch` flag auto-reloads data when the file changes
- **Multiple formats** — load from CSV or JSON
- **Relevance scoring** — results ranked by fuzzy match with percentage scores
- **Graceful shutdown** — clean server stop on SIGINT/SIGTERM
- **Single binary** — zero runtime dependencies, cross-platform

## Quick Start

### Install

```bash
# From source
go install github.com/akhil-datla/athena/cmd/athena@latest

# Or download a binary from Releases
# https://github.com/akhil-datla/Athena/releases
```

### Prepare your data

**CSV** (no header needed):

```csv
How do I reset my password?,Go to Settings > Security > Reset Password.
What are your business hours?,Monday to Friday 9am-5pm EST.
Do you offer refunds?,Yes. Request within 30 days by contacting support.
```

**JSON:**

```json
[
  {"question": "How do I reset my password?", "answer": "Go to Settings > Security > Reset Password."},
  {"question": "What are your hours?", "answer": "Monday to Friday 9am-5pm."}
]
```

### Interactive search (TUI)

```bash
athena search -f faq.csv
```

Opens a live fuzzy finder in your terminal. Type to search, arrow keys to navigate, Enter to select.

### One-shot search

```bash
athena search -f faq.csv "reset password"

#  RELEVANCE  QUESTION                     ANSWER
1  54%        How do I reset my password?  Go to Settings > Security > Reset Password...
```

```bash
athena search -f faq.csv --format json "billing"   # JSON output
athena search -f faq.csv -l 5 "shipping"            # limit to 5 results
```

### Start the server

```bash
athena serve -f faq.csv
athena serve -f faq.csv -p 3000         # custom port
athena serve -f faq.csv --watch          # auto-reload on file changes
```

Opens the web UI at `http://localhost:8080` and the API at `/api/search`.

### Docker

```bash
docker build -t athena .
docker run -v $(pwd)/data:/data -p 8080:8080 athena serve -f /data/faq.csv
```

## API Reference

### `GET /api/search`

| Parameter | Type   | Default | Description              |
|-----------|--------|---------|--------------------------|
| `q`       | string | —       | Search query (required)  |
| `limit`   | int    | 10      | Max results (0 = all)    |

```json
{
  "query": "password",
  "count": 1,
  "results": [
    {
      "question": "How do I reset my password?",
      "answer": "Go to Settings > Security > Reset Password.",
      "score": 8
    }
  ]
}
```

### `GET /api/health`

```json
{"status": "ok"}
```

### `GET /api/stats`

```json
{"entries": 20, "uptime": "1h23m45s"}
```

## Development

```bash
make test     # run tests with race detector
make build    # build binary to bin/
make lint     # run golangci-lint
make run      # build and run with example data
```

## License

[MIT](LICENSE) — Akhil Datla

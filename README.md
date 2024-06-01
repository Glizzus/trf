# Totally Real Facts

Totally Real Facts is a parody website of Snopes.

## Getting Started

### Requirements

- [Docker](https://www.docker.com/)

- [Docker Compose](https://docs.docker.com/compose/)

- [Go](https://golang.org/) (Optional)

- [PostgreSQL](https://www.postgresql.org/) (Optional)

### Running with Docker Compose

```bash
docker-compose up --build
```

If you have Docker Compose 2.22.0 or later, you can develop with hot reload:

```bash
docker compose up --build --watch
```

### Running with Go

Currently, running with Golang is cumbersome, but can be done with the following steps:

1. Install Go dependencies:

    ```bash
    go mod download
    ```

2. Export environment variables:

    ```bash
    # Postgres
    export MINISTRY_POSTGRES_HOST=localhost
    export MINISTRY_POSTGRES_USER=ministry
    export MINISTRY_POSTGRES_PASSWORD=ministry
    export MINISTRY_POSTGRES_DB=ministry

    # Spoofing
    export SPOOFER_TYPE=mock
    ```

3. Run the server:

    ```bash
    go run cmd/ministry/main.go serve
    ```

## Configuration

### Environment Variables

- Postgres

    | Name | Description | Required |
    | --- | --- | --- |
    | `MINISTRY_POSTGRES_HOST` | PostgreSQL host | Yes |
    | `MINISTRY_POSTGRES_USER` | PostgreSQL user | Yes |
    | `MINISTRY_POSTGRES_PASSWORD` | PostgreSQL password | Yes |
    | `MINISTRY_POSTGRES_DB` | PostgreSQL database | Yes |
    | `MINISTRY_POSTGRES_PORT` | PostgreSQL port | No (default: `5432`) |

- Spoofing

    | Name | Description | Required |
    | --- | --- | --- |
    | `MINISTRY_SPOOFER_TYPE` | Type of spoofer to use. Options are `mock`, and `openai` | Yes |
    | `MINISTRY_OPENAI_KEY` | OpenAI API key (required if `MINISTRY_SPOOFER_TYPE` is `openai`) | No |

## Endpoints

### `GET /latest`

- Description: Returns a HTML page cataloging the latest articles.

- Response:
  - Content-Type: `text/html`
  - Body: [Click here to view the full HTML template](./templates/latest.html)

### `GET /{slug}`

- Description: Returns a HTML page for a specific article.

- Response:
  - Content-Type: `text/html`
  - Body: [Click here to view the full HTML template](./templates/article.html)

### `GET /healthz`

- Description: Returns a 200 status code if the server is healthy.

- Response:
  - Status Code: `200`

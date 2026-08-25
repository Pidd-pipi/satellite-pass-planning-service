# Satellite Pass Planning Backend

This directory is the Go module and HTTP service entrypoint.

```bash
go test ./...
go build ./...
go run .
```

The `main` package starts the service on `PORT` (default `8080`). It exposes `GET /healthz`, `GET /api/passes`, `POST /api/passes`, `/`, and `/app.js`.

# Email Service

A Go microservice for handling email operations, featuring Redis-based background processing and structured logging.

## Features

- **Redis Integration**: Uses Redis streams for reliable background email processing.
- **Structured Logging**: Implements `log/slog` for structured logging.
  - **Console**: Colorized, human-readable logs for development.
  - **File**: JSON-formatted logs (`logs/server.log`) for production and analysis.
- **Graceful Shutdown**: Handles OS signals to ensure tasks complete before exiting.
- **Health Check**: Simple `/health` endpoint for readiness probes.

## Project Structure

- `cmd/api`: Application entry point.
- `internal/`: Private application code.
  - `config`: Configuration management using `.env` and environment variables.
  - `handlers`: HTTP handlers/controllers.
  - `models`: Domain models.
  - `repository`: Data access layer (Redis client).
  - `service`: Business logic.
  - `worker`: Background worker logic for processing emails.
  - `logger`: Custom logging utilities (Fanout handler).
- `pkg/`: Public library code.

## Configuration

The application is configured via environment variables. You can set these in a `.env` file in the root directory.

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_ADDR` | Address of the Redis server | `localhost:6379` |
| `SERVER_PORT` | HTTP Port for the API | `:8081` |

## Prerequisites

- **Go**: 1.25+
- **Redis**: Running instance on `localhost:6379` (or configure via `REDIS_ADDR`).

## Running the application

1.  **Start Redis**:
    Ensure Redis is running.
    ```bash
    docker run -p 6379:6379 -d redis
    ```

2.  **Run the service**:
    ```bash
    go run cmd/api/main.go
    ```

3.  **Check logs**:
    Logs differ by output:
    - **Console**: Look for colorized logs in your terminal.
    - **File**: Check `logs/server.log` for JSON logs.

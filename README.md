# Email Service

A Go microservice for handling email operations.

## Project Structure

- `cmd/api`: Application entry point.
- `internal/`: Private application code.
  - `config`: Configuration management.
  - `handlers`: HTTP handlers/controllers.
  - `models`: Domain models.
  - `repository`: Data access layer.
  - `service`: Business logic.
- `pkg/`: Public library code.

## Running the application

```bash
go run cmd/api/main.go
```

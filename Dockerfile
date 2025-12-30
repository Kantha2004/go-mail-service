# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies if needed (none strictly required for basic go build on alpine, but git is good practice)
# RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 is important for alpine runtime compatibility
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Run stage
FROM alpine:latest

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Create logs directory
RUN mkdir -p logs

# Expose port 8081 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./main"]

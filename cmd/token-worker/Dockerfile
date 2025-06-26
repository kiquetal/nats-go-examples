FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /go/bin/token-worker ./cmd/token-worker

# Use a minimal alpine image for the final container
FROM alpine:3.19

# Add CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /go/bin/token-worker /app/token-worker

# Copy configuration file
COPY configs/app.json /app/configs/app.json

# Set working directory
WORKDIR /app

# Command to run
ENTRYPOINT ["/app/token-worker"]
CMD ["-config", "configs/app.json"]

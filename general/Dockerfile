# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder


WORKDIR /app

# Copy go.mod and go.sum to cache dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o custom-scheduler main.go

# Stage 2: Create a minimal production image
FROM alpine

WORKDIR /

# Copy the built binary from the builder stage
COPY --from=builder /app/custom-scheduler /usr/local/bin/custom-scheduler

# Set the entry point to run your custom scheduler
ENTRYPOINT ["/usr/local/bin/custom-scheduler"]
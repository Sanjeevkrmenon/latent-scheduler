FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first, to maximize caching for deps
COPY go.mod go.sum ./
RUN go mod download

# Now copy your Go source code
COPY . .

# Do mod tidy to update go.mod and go.sum as needed (esp. for indirect deps)
RUN go mod tidy

RUN go build -o custom-scheduler main.go

# Minimal runtime image
FROM alpine
WORKDIR /
COPY --from=builder /app/custom-scheduler /usr/local/bin/custom-scheduler
ENTRYPOINT ["/usr/local/bin/custom-scheduler"]
# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24.3-alpine AS builder
WORKDIR /app

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Enable CGO for sqlite3 and set target architecture
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=1
RUN GOARCH=$TARGETARCH GOOS=$TARGETOS go build -o server main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies for sqlite3
RUN apk add --no-cache libstdc++ sqlite-libs

COPY --from=builder /app/server ./
EXPOSE 50051
ENTRYPOINT ["/app/server"]
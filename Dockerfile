# Use official Golang image as builder
FROM golang:1.21 as builder

WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN go build -o grpc-server main.go

# Use minimal image for final stage
FROM gcr.io/distroless/base-debian11

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/grpc-server .
COPY --from=builder /app/database.db .

# Command to run the executable
ENTRYPOINT ["/app/grpc-server"]

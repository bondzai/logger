# Stage 1: Build the application
FROM golang:1.21 AS builder

WORKDIR /go/src/app

# Copy only the necessary files for the Go modules
COPY go.mod .
COPY go.sum .

# Download and install Go dependencies
RUN go mod download
RUN go mod tidy

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o logger ./cmd

# Stage 2: Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /go/src/app/logger /app/logger

# Expose sprcific port to the outside world
EXPOSE 50051

# Command to run the executable
CMD ["./logger"]

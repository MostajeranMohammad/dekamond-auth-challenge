# First stage: Build the Go application
FROM golang:alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -ldflags="-s -w" -o main ./cmd/auth/main.go

# Second stage: Create a lightweight image to run the application
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled application from the builder stage
COPY --from=builder /app/main .

# Copy migration files from the builder stage
COPY --from=builder /app/database ./database

# Expose the application port explicitly (no env dependency)
EXPOSE 8080

# Set the entry point to run the compiled application
CMD ["./main"]
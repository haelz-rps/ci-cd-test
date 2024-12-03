# Stage 1: Build stage
FROM golang:1.20.4-bullseye AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go dependencies
RUN go mod download

# Copy the application source code
COPY . .

ARG cmd=internal
# Build the Go application
RUN go build -o myapp ./cmd/$cmd

# Stage 2: Production-ready stage
FROM debian:bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/myapp .
COPY wait-for-it.sh .
RUN chmod +x wait-for-it.sh

# Expose the port that the application listens on
EXPOSE 8080

# Set the entry point for the container
CMD ["./myapp"]

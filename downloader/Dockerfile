# Build stage
FROM golang:1.22-alpine AS builder

# Set the working directory inside the builder container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o downloader ./cmd

# Final stage
FROM jrottenberg/ffmpeg:4.1-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go application from the builder stage
COPY --from=builder /app/downloader /app/downloader
COPY ./config/config.yaml /app/config/config.yaml
# Set the entrypoint to run your Go application
ENTRYPOINT ["./downloader"]
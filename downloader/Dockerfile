# Build the Go application
FROM golang:latest AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY * ./

RUN go mod download

# Build the Go application
RUN go build -o downloader .

# Use a lightweight FFmpeg image
FROM jrottenberg/ffmpeg:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go application from the builder stage
COPY --from=builder /app/downloader /app/downloader

# Set the entrypoint to run your Go application
ENTRYPOINT ["./downloader"]

# /my-microservice/go-app/Dockerfile

# Use a Go version that matches go.mod (>= 1.24.4)
FROM golang:1.24.4-alpine AS builder

# Install Git (required for go mod download)
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application, statically linked for a smaller final image
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /oms main.go

# Start a new, much smaller image for the final container
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /oms .

# Expose the port the microservice will listen on
EXPOSE 8090

# Command to run the application when the container starts
CMD ["./oms"]
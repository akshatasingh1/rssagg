# Use the official Golang image to build the app
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go application
RUN go build -o rssagg .

# Use a lightweight image to actually run the compiled app
FROM alpine:latest
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/rssagg .
#COPY --from=builder /app/.env . 

# Expose the port your app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./rssagg"]
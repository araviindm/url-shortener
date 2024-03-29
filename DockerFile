# Use an official Golang runtime as the base image
FROM golang:1.20.3-alpine AS builder

# Set the current working directory inside the container
WORKDIR /

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Use a lightweight Alpine image to create a final Docker image
FROM alpine:latest

# Set the current working directory inside the container
WORKDIR /

# Copy the executable from the builder stage
COPY --from=builder /main .

# Expose port 8081
EXPOSE 8081

# Command to run the executable
CMD ["./main"]

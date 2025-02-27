# Use the official Go image
FROM golang:1.22

# Set the working directory
WORKDIR /app

# Copy Go modules and dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port the service runs on
EXPOSE 4002

# Run the application
CMD ["./main"]

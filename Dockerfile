FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app/scr/app

# Copy the Go modules manifests
COPY . .

#Expose the port that the application will run on
EXPOSE 8000

# Build the Go application
RUN go build -o main cmd/main.go

# Run the Go application
CMD ["./main"]
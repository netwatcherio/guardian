# Use an official Go runtime as a parent image
FROM golang:1.21.1-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace.
COPY . /app

# Download all the dependencies
RUN go mod download

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Set environment variables
ENV DEBUG=true
ENV LISTEN=0.0.0.0:8080
ENV MONGO_URI=""
ENV MAIN_DB=""
ENV KEY=""

# Command to run the executable
CMD ["/app/main"]

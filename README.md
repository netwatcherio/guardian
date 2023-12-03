# Guardian NetWatcher

## Overview

Guardian NetWatcher is the backend component of the NetWatcher suite. It facilitates communication between the frontend client and the metrics collecting agent, storing data in MongoDB. The system uses probes to collect various metrics like MTR, ping, rperf (simulated traffic), which the client uses to generate graphs and other visualizations.

## Environment Variables

The following environment variables are required for running the application:

LISTEN=[ip_address:port](ip_address:port)
MONGO\_URI=<mongodb\_connection\_string>
MAIN_DB=<database\_name>
KEY=<your\_secret\_key>```


**Note**: Replace the values with your actual configuration. Do not use example values in production.

## Docker Compose Setup

Here's an example of a Docker Compose setup for the Guardian NetWatcher:

```yaml

version: '3.8'
services:
  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
    restart: unless-stopped

  nw-guardian:
    image: nw-guardian
    environment:
      DEBUG: "true"
      LISTEN: "0.0.0.0:8080"
      MONGO_URI: "mongodb://<username>:<password>@<your_mongodb_host>:27017/<database_name>"
      MAIN_DB: "netwatcher"
      KEY: "<your_secret_key>"
    depends_on:
      - mongodb

  nw-client:
    image: nw-client
    environment:
      NW_GLOBAL_ENDPOINT: "https://api.netwatcher.io"
    ports:
      - "3000:3000"

  mongodb:
    image: mongo:4.4.6
    container_name: mongodb
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: <username>
      MONGO_INITDB_ROOT_PASSWORD: <password>
    volumes:
      - mongodb_data:/data/db

volumes:
  mongodb_data:
```

**Note**: Replace `<username>`, `<password>`, `<your_mongodb_host>`, and `<database_name>` with your actual MongoDB credentials and details.


## Building the Guardian Docker Image

Here's a Dockerfile for building the Guardian NetWatcher backend service:

```
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

# Set default environment variables (can be overridden)
ENV DEBUG=true
ENV LISTEN=0.0.0.0:8080
ENV MONGO_URI=""
ENV MAIN_DB=""
ENV KEY=""

# Command to run the executable
CMD ["/app/main"]
```


## Caddy Configuration

Here's an example Caddyfile configuration for the NetWatcher services:

```
api.netwatcher.io {
    reverse_proxy http://nw-guardian:8080
}

app.netwatcher.io {
    @ws {
	header Connection *Upgrade*
	header Upgrade websocket
    }
    reverse_proxy http://nw-client:3000
}

```



## License

[`GNU Affero General Public License v3.0`

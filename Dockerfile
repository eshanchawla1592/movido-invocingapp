FROM golang:1.20.0-alpine3.17

# Set working directory
WORKDIR /app

# Copy source code
COPY . ./

# Get dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /invoicinig-app

# Expose port
EXPOSE 8080

# Run
CMD ["/invoicinig-app"]
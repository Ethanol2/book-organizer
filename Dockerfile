# Step 1 ======================

# Build the frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend

# Copy dependencies and build
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Step 2 ======================

# Build the Go application
# Using alpine instead of debian because alpine is smaller and provides less attack surface
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependencies to improve build caching speeds
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the source code
COPY backend/ .

# Build the Go application
# CGO_ENABLED=0 creates a static self contained binary with no dependencies
# GOOS=linux ensures it targets Linux architecture
RUN CGO_ENABLED=0 GOOS=linux go build -o ./book-organizer-server .


# Step 3 ======================

# Build the runtine container
# Not using :latest to avoid unexpected behaviour from new updates
FROM alpine:3.19

# Security best practice: Creates a non-root system user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Sets the /app folder as the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/book-organizer-server ./book-organizer-server

# Copy the compiled vue files to the backend
COPY --from=frontend-builder /frontend/dist ./static

# Ensure the app user owns the binary and folders
RUN mkdir ./data
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the port and run the server
EXPOSE 8080
CMD ["./book-organizer-server"]

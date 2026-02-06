# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/src ./src
COPY frontend/svelte.config.js frontend/vite.config.js ./
# Copy tailwind config if it exists, otherwise ignore
# For now, just copy config files we know exist.
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the binary
RUN go build -o jjmc cmd/jjmc/main.go

# Stage 3: Runtime
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies (git for BuildTools, java for Minecraft)
RUN apk add --no-cache \
    git \
    openjdk21 \
    bash \
    curl

# Copy artifacts
COPY --from=backend-builder /app/jjmc .
COPY --from=frontend-builder /app/build ./frontend/build
COPY --from=frontend-builder /app/package.json ./frontend/
COPY templates ./templates

# Expose port
EXPOSE 3001 2024

# Create data directories
RUN mkdir -p data/instances data/backups .tools

# Volume for persistent data
VOLUME ["/app/data"]

CMD ["./jjmc"]

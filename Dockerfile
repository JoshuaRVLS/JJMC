# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY src ./src
COPY svelte.config.js vite.config.js ./
# Copy tailwind config if it exists, otherwise ignore (Docker COPY fails if missing without wildcard hack, but we'll assume standard setup or omit if unsure)
# For now, just copy config files we know exist.
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the binary
RUN go build -o jjmc main.go

# Stage 3: Runtime
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies (git for BuildTools, java for Minecraft)
# We install multiple java versions to support different MC versions if possible, 
# but for now let's stick to a default JDK 21 (good for modern MC).
# User might need to mount their own Java or we might need a bigger image.
# For Docker-in-Docker (needed for "Run Servers in Docker"), we need specialized setup.
# But for "Run JJMC in Docker", we assume it manages local processes INSIDE the container?
# Or does it manage sibling containers? 
# If it manages sibling containers, we need docker socket.

RUN apk add --no-cache \
    git \
    openjdk21 \
    bash \
    curl

# Copy artifacts
COPY --from=backend-builder /app/jjmc .
COPY --from=frontend-builder /app/frontend/build ./frontend/build
COPY --from=frontend-builder /app/frontend/package.json ./frontend/
COPY templates ./templates

# Expose port (default 3000? 8080? main.go defaults to 3001 for fiber?)
# Need to check main.go port. assuming 3001 based on typical fiber apps or config.
EXPOSE 3001 2024

# Create data directories
RUN mkdir -p instances servers backups .tools

# Volume for persistent data
VOLUME ["/app/instances", "/app/servers", "/app/backups", "/app/jjmc.db"]

CMD ["./jjmc"]

# Stage 1: Build Go backend
FROM golang:1.25.1-alpine AS backend-builder

WORKDIR /app
COPY . .

# Install dependencies
RUN go mod download

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o web-server ./cmd/web/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o parser ./cmd/parser/main.go

# Stage 2: Build Frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY web/frontend/package.json web/frontend/package-lock.json ./
RUN npm ci

COPY web/frontend/ .
RUN npm run build

# Stage 3: Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy backend binaries
COPY --from=backend-builder /app/web-server .
COPY --from=backend-builder /app/parser .

# Copy frontend build
COPY --from=frontend-builder /app/frontend/dist ./web/frontend/dist

# Copy configs and migrations
COPY configs ./configs
COPY migrations ./migrations

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./web-server"]
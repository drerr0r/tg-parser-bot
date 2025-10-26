# Stage 1: Build Go backend
FROM golang:1.25.1-alpine AS backend-builder

WORKDIR /app
COPY . .

RUN go mod download

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

# Copy frontend build - ИСПРАВЛЯЕМ ПУТЬ!
COPY --from=frontend-builder /app/frontend/dist ./web/frontend/dist

# Copy configs and migrations
COPY configs ./configs
COPY migrations ./migrations

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

CMD ["./web-server"]
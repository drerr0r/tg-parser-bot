FROM golang:1.25.1-alpine

RUN apk add --no-cache nodejs npm

WORKDIR /app
COPY . .

# Удаляем dev config
RUN rm -f configs/config.yaml

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o web-server ./cmd/web/main.go

# Детальная проверка сборки фронтенда
RUN echo "=== Checking frontend structure ==="
RUN ls -la web/frontend/
RUN echo "=== Checking package.json ==="
RUN cat web/frontend/package.json
RUN echo "=== Installing dependencies ==="
RUN cd web/frontend && npm ci --verbose
RUN echo "=== Building frontend ==="
RUN cd web/frontend && npm run build --verbose
RUN echo "=== Verifying build ==="
RUN find web/frontend/dist -type f | head -20 || echo "=== BUILD FAILED ==="

EXPOSE 8080

CMD ["./web-server"]
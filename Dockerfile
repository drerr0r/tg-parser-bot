FROM golang:1.25.1-alpine

# Устанавливаем Node.js и npm
RUN apk add --no-cache nodejs npm

WORKDIR /app
COPY . .

# ПРОСТО УДАЛЯЕМ config.yaml - приложение будет использовать env variables
RUN rm -f configs/config.yaml

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o web-server ./cmd/web/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o parser ./cmd/parser/main.go

RUN cd web/frontend && npm ci && npm run build

RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

CMD ["./web-server"]
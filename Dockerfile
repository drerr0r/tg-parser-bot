FROM golang:1.25.1-alpine

RUN apk add --no-cache nodejs npm

WORKDIR /app
COPY . .

# Удаляем dev config
RUN rm -f configs/config.yaml

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o web-server ./cmd/web/main.go

RUN cd web/frontend && npm ci && npm run build

EXPOSE 8080

CMD ["./web-server"]
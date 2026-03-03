FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY configs /app/configs
ENV CONFIG_PATH=/app/configs/config.yaml
EXPOSE 8080
CMD ["/app/server"]

FROM golang:1.26-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/ecommerce

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]

FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o subscription-manager ./cmd/app

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/subscription-manager .
EXPOSE 8080
CMD ["./subscription-manager", "serve"]

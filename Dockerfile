FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloudflare-ingress-controller ./cmd/controller

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/cloudflare-ingress-controller .

CMD ["./cloudflare-ingress-controller"]
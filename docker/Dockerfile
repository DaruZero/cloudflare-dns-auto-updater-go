FROM docker.io/golang:1.20.1-alpine3.17 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY cmd/ ./

RUN GOOS=linux go build -o app ./...

FROM scratch

COPY --from=builder /build/app /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["./app"]
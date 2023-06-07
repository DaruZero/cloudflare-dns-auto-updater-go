FROM docker.io/golang:1.20.5-alpine3.17 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY cmd/ ./cmd/

RUN GOOS=linux go build -o build/ ./...

FROM scratch

COPY --from=builder /build/build/cmd /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["./app"]
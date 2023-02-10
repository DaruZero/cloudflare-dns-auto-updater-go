FROM docker.io/golang:1.20.0-alpine3.17 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY cmd/ ./

RUN GOOS=linux go build -o app ./...

FROM scratch

COPY --from=builder /build/app /app

CMD ["./app"]
FROM docker.io/golang:1.20.7-alpine3.17 as builder

WORKDIR /go/src/github.com/daruzero/cloudflare-dns-autoupdater-go

COPY . .

RUN GOOS=linux go build -o app ./cmd

RUN chmod +x app

FROM scratch

COPY --from=builder /go/src/github.com/daruzero/cloudflare-dns-autoupdater-go/app /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["./app"]
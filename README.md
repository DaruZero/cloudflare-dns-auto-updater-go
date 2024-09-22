# Cloudflare DNS IP Updater - Go version

[![Go Report Card](https://goreportcard.com/badge/github.com/DaruZero/cloudflare-dns-auto-updater-go)](https://goreportcard.com/report/github.com/DaruZero/cloudflare-dns-auto-updater-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/daruzero/cfautoupdater-go)](https://hub.docker.com/r/daruzero/cfautoupdater-go)
[![Docker Image Size (tag)](https://img.shields.io/docker/image-size/daruzero/cfautoupdater-go/latest)](https://hub.docker.com/r/daruzero/cfautoupdater-go)

If you have some self-hosted services exposed to the internet but not a static public IP, you certainly faced the annoying task to access the CloudFlare dashboard and manually change all your records with the new IP everytime it changes.

What if I tell you that it can be automated? With this simple application you just have to spin a Docker container and not worry about your IP changing anymore.

It will continuously run and fetch your public IP at a given interval, detecting if it changes and sending a request to the CloudFlare API to update you records. You can even choose to be notified by email when that happens.

This is based on the [daruzero/cloudflare-dns-auto-updater](https://github.com/DaruZero/cloudflare-dns-auto-updater) python script, but it has been rewritten in Go.

> ⚠️ Be aware, this only works if you are using the Cloudflare DNS.

## Requirements

- A CloudFlare account
- Cloudflare Global API Key
- The domain name you want to change the record of
- (optional) The ID of the A record you want to change ([how to](https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records))

## Installation

- Plain Docker

  ```shell
  docker run -d \
    -e EMAIL=<YOUR_CF_LOGIN_EMAIL> \
    -e AUTH_KEY=<YOUR_API_KEY> \
    -e ZONE_NAME=<YOUR_ZONE_NAME> \
    daruzero/cfautoupdater-go:latest
  ```

- Docker Compose (see <code><a href="./build/.env.example">.env.example</a></code> for the env file)

  ```yaml
  version: '3.8'

  services:
    app:
      image: daruzero/cfautoupdater-go:latest
      env_file: .env
      restart: unless-stopped
  ```

### Environment variables

#### Required

| Variable    | Example value                                 | Description                                                                                                               |
|-------------|-----------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|
| `EMAIL`     | <johndoe@example.com>                           | Email address of your CloudFlare account                                                                     |
| `AUTH_KEY`  | c2547eb745079dac9320b638f5e225cf483cc5cfdda41 | CloudFlare Global API Key                                                                                            |
| `ZONE_NAME` | example.com                                   | Comma separated list of domains to update. |
| `ZONE_ID`   | 372e67954025e0ba6aaa6d586b9e0b59              | Comma separated list of domain IDs to update     |

> **Note:**
>
> - You can set either `ZONE_ID` or `ZONE_NAME`. If you specify both, the value of `ZONE_ID` takes precedence.
>

#### Optional

| Variable           | Example value                    | Description                                                                                                                                | Default |
|--------------------|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `RECORD_ID`        | 372e67954025e0ba6aaa6d586b9e0b59 | The ID of the record you want to change. Leave blank to update all records of the zone.                                                    | -       |
| `CHECK_INTERVAL`   | 86400                            | The amount of seconds the script should wait between checks                                                                                | `86400` |
| `NOTIFICATION_URLS` | smtp://username:password@host:587/?fromAddress=<sender@example.com>&toAddresses=<recipient@example.com>              | Comma separated list of URLs as consumed by [`shoutrrr`](https://containrrr.dev/shoutrrr/v0.8/services/overview/)  | -       |

---

## Future implementation

- [x] Possibility to update multiple domains
- [x] Support for other SMTP servers other than Google's
- [x] Support for other notification systems
- [ ] Support for other DNS services

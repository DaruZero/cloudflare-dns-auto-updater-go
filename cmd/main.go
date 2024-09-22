package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/dnsapi"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/logger"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/notifier"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("LOG_LEVEL")
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	zap.S().Info("Starting Cloudflare CFDNS Auto Updater")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	currentIpChan := make(chan string)
	go getCurrentIp(ctx, currentIpChan)
	lastIp := ""

	cfg, err := config.New()
	if err != nil {
		zap.S().Fatal(err)
	}

	dns, err := dnsapi.New(cfg)
	if err != nil {
		zap.S().Fatal(err)
	}

	notify := notifier.New(cfg)

	for {
		select {
		case ip := <-currentIpChan:
			if ip != lastIp {
				zap.S().Infof("New IP detected: %s", ip)
				lastIp = ip
				wg.Add(1)
				go func(ip string) {
					defer wg.Done()
					if updatedRecords, err := dns.UpdateRecords(ip); err == nil {
						if notify == nil {
							zap.S().Debug("Skip sending notification")
							return
						}
						wg.Add(1)
						go func(updatedRecords map[string][]string) {
							defer wg.Done()
							notify.Send(updatedRecords, ip)
						}(updatedRecords)
					}
				}(ip)
			}
		case <-ctx.Done():
			zap.S().Info("Shutting down...")
			wg.Wait()
			return
		}
	}
}

// getCurrentIp fetches the current public ip address every second,
// and sends it to the currentIpChan channel
func getCurrentIp(ctx context.Context, currentIpChan chan<- string) {
	url := "https://api.ipify.org"
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		zap.S().Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				zap.S().Fatal(err)
			}

			if res.StatusCode != http.StatusOK {
				zap.S().Error("Error getting current ip. Status code: %d", res.StatusCode)
				continue
			}

			var bodyString string
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				zap.S().Fatal(err)
			}
			bodyString = string(bodyBytes)
			res.Body.Close()

			currentIpChan <- bodyString
		}
	}
}

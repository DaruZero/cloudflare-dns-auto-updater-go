package main

import (
	"time"

	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/dnsapi"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/logger"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/notifier"
	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger("LOG_LEVEL")
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	zap.S().Info("Starting Cloudflare CFDNS Auto Updater")

	cfg := config.NewConfig()

	dns := dnsapi.NewDNS(cfg)

	var notify *notifier.Notifier
	if cfg.SenderAddress != "" && cfg.SenderPassword != "" && cfg.ReceiverAddress != "" {
		notify = notifier.New(cfg)
	}

	for {
		updatedRecords := dns.UpdateRecords()
		if len(updatedRecords) > 0 {
			zap.S().Infof("Updated %d records", len(updatedRecords))
			if notify != nil {
				err := notify.SendEmail(updatedRecords, dns.CurrentIP)
				if err != nil {
					zap.S().Errorf("Error sending email: %s", err)
				}
			}
		} else {
			zap.S().Info("No records updated")
		}

		zap.S().Infof("Sleeping for %d seconds", cfg.CheckInterval)
		time.Sleep(time.Duration(cfg.CheckInterval) * time.Second)

		dns.GetCurrentIP()
		dns.GetRecords()
	}
}

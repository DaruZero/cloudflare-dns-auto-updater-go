package main

import (
	"cloudflare-dns-auto-updater-go/cmd/config"
	"cloudflare-dns-auto-updater-go/cmd/dnsapi"
	"cloudflare-dns-auto-updater-go/cmd/notification"
	"cloudflare-dns-auto-updater-go/cmd/utils"
	"go.uber.org/zap"
	"time"
)

func main() {
	log := utils.NewLogger("LOG_LEVEL")
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	zap.S().Info("Starting Cloudflare CFDNS Auto Updater")

	cfg := config.NewConfig()

	dns := dnsapi.NewDNS(cfg)

	var notifier *notification.Notifier
	if cfg.SenderAddress != "" && cfg.SenderPassword != "" && cfg.ReceiverAddress != "" {
		notifier = notification.NewNotifier(cfg)
	}

	for {
		updatedRecords, updated := dns.UpdateRecords()
		if updated {
			zap.S().Infof("Updated %d records", len(updatedRecords))
			if notifier != nil {
				err := notifier.SendEmail(updatedRecords, dns.CurrentIP)
				if err != nil {
					zap.S().Errorf("Error sending email: %s", err)
				}
			}
		} else {
			zap.S().Info("No records updated")
		}

		zap.S().Infof("Sleeping for %d seconds", cfg.CheckInterval)
		time.Sleep(time.Duration(cfg.CheckInterval) * time.Second)

		dns.CurrentIP = dns.GetCurrentIP()
		dns.Records = dns.GetRecords()
	}
}

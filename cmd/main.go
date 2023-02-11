package main

import (
	"go.uber.org/zap"
	"time"
)

func main() {
	log := NewLogger("LOG_LEVEL")
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	zap.S().Info("Starting Cloudflare DNS Auto Updater")

	cfg := NewConfig()

	dns := NewDns(cfg)

	var notifier *Notifier
	if cfg.SenderAddress != "" && cfg.SenderPassword != "" && cfg.ReceiverAddress != "" {
		notifier = NewNotifier(cfg)
	}

	for {
		updatedRecords, updated := dns.UpdateRecords()
		if updated {
			zap.S().Infof("Updated %d records", len(updatedRecords))
			if notifier != nil {
				err := notifier.SendEmail(updatedRecords, dns.CurrentIp)
				if err != nil {
					zap.S().Errorf("Error sending email: %s", err)
				}
			}
		} else {
			zap.S().Info("No records updated")
		}

		zap.S().Infof("Sleeping for %d seconds", cfg.CheckInterval)
		time.Sleep(time.Duration(cfg.CheckInterval) * time.Second)

		dns.CurrentIp = dns.GetCurrentIp()
		dns.Records = dns.GetRecords()
	}
}

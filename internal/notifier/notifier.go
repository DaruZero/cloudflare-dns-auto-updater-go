package notifier

import (
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"go.uber.org/zap"
)

type Notifier struct {
	Sender *router.ServiceRouter
}

// New creates a new Notifier
func New(cfg *config.Config) *Notifier {
	zap.S().Debug("Creating notifier")

	if len(cfg.NotificationURLs) == 0 {
		zap.S().Info("Notification URL not found, skipping.")
		return nil
	}

	sender, err := shoutrrr.CreateSender(cfg.NotificationURLs...)
	if err != nil {
		zap.S().Error(err)
		return nil
	}

	return &Notifier{
		Sender: sender,
	}
}

func (n *Notifier) Send(updatedRecords map[string][]string, newIP string) {
	zap.S().Info("Sending notifications")

	errors := n.Sender.Send(formatMessage(updatedRecords, newIP), &types.Params{"title": "Public IP address changed"})
	for _, err := range errors {
		zap.S().Error(err)
	}
}

func formatMessage(updatedRecords map[string][]string, newIP string) string {
	msg := "Your IP address has changed to " + newIP + " for the following record(s):\r\n"

	for zone, records := range updatedRecords {
		msg = msg + zone + "\r\n"
		for _, record := range records {
			msg = msg + "\t- " + record + "\r\n"
		}
	}

	return msg
}

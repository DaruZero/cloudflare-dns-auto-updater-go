package notifier

import (
	"net/smtp"

	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/config"
	"go.uber.org/zap"
)

type Notifier struct {
	Email Email
}

type Email struct {
	SenderAddress   string
	SenderPassword  string
	ReceiverAddress string
	SMTPServer      string
	SMTPPort        string
}

// New creates a new Notifier
func New(cfg *config.Config) *Notifier {
	zap.S().Debug("Creating notifier")
	return &Notifier{
		Email: Email{
			SenderAddress:   cfg.SenderAddress,
			SenderPassword:  cfg.SenderPassword,
			ReceiverAddress: cfg.ReceiverAddress,
			SMTPServer:      "smtp.gmail.com",
			SMTPPort:        "587",
		},
	}
}

// SendEmail sends an email notification
func (n *Notifier) SendEmail(updatedRecords map[string][]string, newIP string) error {
	zap.S().Info("Sending email notification")
	auth := smtp.PlainAuth("", n.Email.SenderAddress, n.Email.SenderPassword, n.Email.SMTPServer)

	to := []string{n.Email.ReceiverAddress}
	msg := []byte("To: " + n.Email.ReceiverAddress + "\r\n" + "Subject: Public IP Address Changed\r\n" + "\r\n" + "Your IP address has changed to " + newIP + " for the following record(s):" + "\r\n")

	for zone, records := range updatedRecords {
		msg = append(msg, []byte(zone+"\r\n")...)
		for _, record := range records {
			msg = append(msg, []byte("\t- "+record+"\r\n")...)
		}
	}

	err := smtp.SendMail(n.Email.SMTPServer+":"+n.Email.SMTPPort, auth, n.Email.SenderAddress, to, msg)
	if err != nil {
		return err
	}

	zap.S().Info("Email notification sent")

	return nil
}

package main

import (
	"go.uber.org/zap"
	"net/smtp"
)

type Notifier struct {
	Email Email
}

type Email struct {
	SenderAddress   string
	SenderPassword  string
	ReceiverAddress string
	SmtpServer      string
	SmtpPort        string
}

// NewNotifier creates a new Notifier
func NewNotifier(cfg *Config) *Notifier {
	zap.S().Info("Creating notifier")
	return &Notifier{
		Email: Email{
			SenderAddress:   cfg.SenderAddress,
			SenderPassword:  cfg.SenderPassword,
			ReceiverAddress: cfg.ReceiverAddress,
			SmtpServer:      "smtp.gmail.com",
			SmtpPort:        "587",
		},
	}
}

// SendEmail sends an email notification
func (n *Notifier) SendEmail(updatedRecords []string, newIp string) error {
	zap.S().Info("Sending email notification")
	auth := smtp.PlainAuth("", n.Email.SenderAddress, n.Email.SenderPassword, n.Email.SmtpServer)

	to := []string{n.Email.ReceiverAddress}
	msg := []byte("To: " + n.Email.ReceiverAddress + "\r\n" + "Subject: Public IP Address Changed\r\n" + "\r\n" + "Your IP address has changed to " + newIp + " for the following record(s):" + "\r\n")
	for _, record := range updatedRecords {
		msg = append(msg, []byte("- "+record+"\r\n")...)
	}

	err := smtp.SendMail(n.Email.SmtpServer+":"+n.Email.SmtpPort, auth, n.Email.SenderAddress, to, msg)
	if err != nil {
		return err
	}

	zap.S().Info("Email notification sent")

	return nil
}

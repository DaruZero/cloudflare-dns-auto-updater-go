package config

import (
	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/utils"
	"go.uber.org/zap"
)

type Config struct {
	AuthKey         string
	CheckInterval   int
	Email           string
	ReceiverAddress string
	RecordID        string
	SenderAddress   string
	SenderPassword  string
	ZoneID          string
	ZoneName        string
}

func NewConfig() *Config {
	zap.S().Info("Loading configuration")
	config := &Config{
		AuthKey:         utils.GetEnv("AUTH_KEY", true, ""),
		CheckInterval:   utils.GetEnvAsInt("CHECK_INTERVAL", false, 86400),
		Email:           utils.GetEnv("EMAIL", true, ""),
		ReceiverAddress: utils.GetEnv("RECEIVER_ADDRESS", false, ""),
		RecordID:        utils.GetEnv("RECORD_ID", false, ""),
		SenderAddress:   utils.GetEnv("SENDER_ADDRESS", false, ""),
		SenderPassword:  utils.GetEnv("SENDER_PASSWORD", false, ""),
		ZoneID:          utils.GetEnv("ZONE_ID", false, ""),
		ZoneName:        utils.GetEnv("ZONE_NAME", false, ""),
	}
	zap.S().Debug("Config loaded")

	if config.ZoneID == "" && config.ZoneName == "" {
		zap.S().Fatal("Either ZONE_ID or ZONE_NAME is required")
	}

	return config
}

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
	RecordIDs       []string
	SenderAddress   string
	SenderPassword  string
	ZoneIDs         []string
	ZoneNames       []string
}

func NewConfig() *Config {
	zap.S().Info("Loading configuration")
	config := &Config{
		AuthKey:         utils.GetEnv("AUTH_KEY", true, ""),
		CheckInterval:   utils.GetEnvAsInt("CHECK_INTERVAL", false, 86400),
		Email:           utils.GetEnv("EMAIL", true, ""),
		ReceiverAddress: utils.GetEnv("RECEIVER_ADDRESS", false, ""),
		RecordIDs:       utils.GetEnvAsStringSlice("RECORD_ID", false, []string{}),
		SenderAddress:   utils.GetEnv("SENDER_ADDRESS", false, ""),
		SenderPassword:  utils.GetEnv("SENDER_PASSWORD", false, ""),
		ZoneIDs:         utils.GetEnvAsStringSlice("ZONE_ID", false, []string{}),
		ZoneNames:       utils.GetEnvAsStringSlice("ZONE_NAME", false, []string{}),
	}
	zap.S().Debug("Config loaded")

	if len(config.ZoneIDs) == 0 && len(config.ZoneNames) == 0 {
		zap.S().Fatal("Either ZONE_ID or ZONE_NAME is required")
	}

	return config
}

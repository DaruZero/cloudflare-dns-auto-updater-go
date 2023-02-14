package main

import (
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
		AuthKey:         GetEnv("AUTH_KEY", true, ""),
		CheckInterval:   GetEnvAsInt("CHECK_INTERVAL", false, 86400),
		Email:           GetEnv("EMAIL", true, ""),
		ReceiverAddress: GetEnv("RECEIVER_ADDRESS", false, ""),
		RecordID:        GetEnv("RECORD_ID", false, ""),
		SenderAddress:   GetEnv("SENDER_ADDRESS", false, ""),
		SenderPassword:  GetEnv("SENDER_PASSWORD", false, ""),
		ZoneID:          GetEnv("ZONE_ID", false, ""),
		ZoneName:        GetEnv("ZONE_NAME", false, ""),
	}
	zap.S().Debug("Config loaded")

	if config.ZoneID == "" && config.ZoneName == "" {
		zap.S().Fatal("Either ZONE_ID or ZONE_NAME is required")
	}

	return config
}

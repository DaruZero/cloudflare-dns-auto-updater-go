package main

import (
	"go.uber.org/zap"
)

type Config struct {
	AuthKey         string
	CheckInterval   int
	Email           string
	ReceiverAddress string
	RecordId        string
	SenderAddress   string
	SenderPassword  string
	ZoneId          string
	ZoneName        string
}

func NewConfig() *Config {
	zap.S().Info("Loading configuration")
	config := &Config{
		AuthKey:         GetEnv("AUTH_KEY", true, ""),
		CheckInterval:   GetEnvAsInt("CHECK_INTERVAL", false, 86400),
		Email:           GetEnv("EMAIL", true, ""),
		ReceiverAddress: GetEnv("RECEIVER_ADDRESS", false, ""),
		RecordId:        GetEnv("RECORD_ID", false, ""),
		SenderAddress:   GetEnv("SENDER_ADDRESS", false, ""),
		SenderPassword:  GetEnv("SENDER_PASSWORD", false, ""),
		ZoneId:          GetEnv("ZONE_ID", false, ""),
		ZoneName:        GetEnv("ZONE_NAME", false, ""),
	}
	zap.S().Debugf("Config loaded: %+v", config)

	if config.ZoneId == "" && config.ZoneName == "" {
		zap.S().Fatal("Either ZONE_ID or ZONE_NAME is required")
	}

	return config
}

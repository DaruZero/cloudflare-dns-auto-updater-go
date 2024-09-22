package config

import (
	"errors"

	"github.com/daruzero/cloudflare-dns-auto-updater-go/pkg/env"
	"go.uber.org/zap"
)

type Config struct {
	AuthKey          string
	Email            string
	NotificationURLs []string
	RecordIDs        []string
	ZoneIDs          []string
	ZoneNames        []string
	CheckInterval    int
}

func New() (config *Config, err error) {
	zap.S().Info("Loading configuration")
	config = &Config{
		AuthKey:          env.GetEnv("AUTH_KEY", true, ""),
		CheckInterval:    env.GetEnvAsInt("CHECK_INTERVAL", false, 86400),
		Email:            env.GetEnv("EMAIL", true, ""),
		RecordIDs:        env.GetEnvAsStringSlice("RECORD_ID", false, []string{}),
		ZoneIDs:          env.GetEnvAsStringSlice("ZONE_ID", false, []string{}),
		ZoneNames:        env.GetEnvAsStringSlice("ZONE_NAME", false, []string{}),
		NotificationURLs: env.GetEnvAsStringSlice("NOTIFICATION_URLS", false, []string{}),
	}
	zap.S().Debug("Config loaded")

	if len(config.ZoneIDs) == 0 && len(config.ZoneNames) == 0 {
		return config, errors.New("no zone ids or zone names provided")
	}

	return config, nil
}

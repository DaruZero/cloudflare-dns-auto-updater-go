package dnsapi

import (
	"context"
	"errors"

	"github.com/cloudflare/cloudflare-go"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/pkg/utils"
	"go.uber.org/zap"
)

type DNSAPI interface {
	UpdateRecords(currentIP string) (updatedRecords map[string][]string, err error)
	getRecords() (err error)
}

// New creates a new Dns struct instance
func New(cfg *config.Config) (dns DNSAPI, err error) {
	zap.S().Debug("Creating new Dns struct")
	dns = &CloudflareAPI{
		Cfg: cfg,
	}

	// Authenticate the APIs
	dns.(*CloudflareAPI).Api, err = cloudflare.New(dns.(*CloudflareAPI).Cfg.AuthKey, dns.(*CloudflareAPI).Cfg.Email)
	if err != nil {
		return nil, err
	}

	// Fetch all zones
	allZones, err := dns.(*CloudflareAPI).Api.ListZones(context.Background())
	if err != nil {
		return nil, err
	}

	// Filter only the selected zones
	for _, zone := range allZones {
		if utils.StringInSlice(zone.ID, dns.(*CloudflareAPI).Cfg.ZoneIDs) || utils.StringInSlice(zone.Name, dns.(*CloudflareAPI).Cfg.ZoneNames) {
			dns.(*CloudflareAPI).Zones = append(dns.(*CloudflareAPI).Zones, zone)
		}
	}

	if len(dns.(*CloudflareAPI).Zones) == 0 {
		return nil, errors.New("no zones found")
	}

	// Get the zones records
	dns.(*CloudflareAPI).Records = make(map[string][]cloudflare.DNSRecord)
	err = dns.getRecords()
	if err != nil {
		return dns, err
	}

	return dns, nil
}

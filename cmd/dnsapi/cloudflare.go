package dnsapi

import (
	"context"
	"errors"

	"github.com/cloudflare/cloudflare-go"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/pkg/utils"
	"go.uber.org/zap"
)

type CFDNS struct {
	Api     *cloudflare.API
	Cfg     *config.Config
	Records map[string][]cloudflare.DNSRecord
	Zones   []cloudflare.Zone
}

// New creates a new Dns struct instance
func New(cfg *config.Config) (dns *CFDNS, err error) {
	zap.S().Debug("Creating new Dns struct")
	dns = &CFDNS{
		Cfg: cfg,
	}

	// Authenticate the APIs
	dns.Api, err = cloudflare.New(dns.Cfg.AuthKey, dns.Cfg.Email)
	if err != nil {
		return nil, err
	}

	// Fetch all zones
	allZones, err := dns.Api.ListZones(context.Background())
	if err != nil {
		return nil, err
	}

	// Filter only the selected zones
	for _, zone := range allZones {
		if utils.StringInSlice(zone.ID, dns.Cfg.ZoneIDs) || utils.StringInSlice(zone.Name, dns.Cfg.ZoneNames) {
			dns.Zones = append(dns.Zones, zone)
		}
	}

	if len(dns.Zones) == 0 {
		return nil, errors.New("no zones found")
	}

	// Get the zones records
	dns.Records = make(map[string][]cloudflare.DNSRecord)
	err = dns.getRecords()
	if err != nil {
		return dns, err
	}

	return dns, nil
}

// getRecords gets all the records for the zone
func (dns *CFDNS) getRecords() (err error) {
	zap.S().Info("Getting records")
	for _, zone := range dns.Zones {
		zap.S().Debugf("%+v", zone)
		records, _, err := dns.Api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zone.ID), cloudflare.ListDNSRecordsParams{Type: "A"})
		if err != nil {
			return err
		}

		if len(dns.Cfg.RecordIDs) != 0 {
			for _, record := range records {
				if utils.StringInSlice(record.ID, dns.Cfg.RecordIDs) {
					dns.Records[zone.ID] = append(dns.Records[zone.ID], record)
				}
			}
		} else {
			dns.Records[zone.ID] = append(dns.Records[zone.ID], records...)
		}
	}

	if len(dns.Records) == 0 {
		return errors.New("no records found")
	}

	return nil
}

// UpdateRecords updates the records with the current ip
func (dns *CFDNS) UpdateRecords(currentIP string) (updatedRecords map[string][]string, err error) {
	if err := dns.getRecords(); err != nil {
		return nil, err
	}

	zap.S().Info("Checking records")
	updatedRecords = make(map[string][]string)

	for zoneID, records := range dns.Records {
		for i, record := range records {
			zap.S().Infof("Updating record %s", record.Name)

			newRecord, err := dns.Api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{ID: record.ID, Content: currentIP})
			if err != nil {
				return nil, err
			}

			records[i] = newRecord

			var zoneName string
			for _, zone := range dns.Zones {
				if zone.ID == zoneID {
					zoneName = zone.Name
				}
			}

			updatedRecords[zoneName] = append(updatedRecords[zoneName], record.Name)
		}
	}

	return updatedRecords, nil
}

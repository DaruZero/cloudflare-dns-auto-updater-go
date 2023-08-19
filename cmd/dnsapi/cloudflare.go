package dnsapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/pkg/utils"
	"go.uber.org/zap"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type CFDNS struct {
	HTTPClient HTTPClient
	Cfg        *config.Config
	Records    map[string][]Record
	Zones      []Zone
}

type Record struct {
	Content  string `json:"content"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type Message struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// New creates a new Dns struct instance
func New(cfg *config.Config) *CFDNS {
	zap.S().Debug("Creating new Dns struct")
	dns := &CFDNS{
		Cfg: cfg,
	}

	dns.HTTPClient = http.DefaultClient

	if len(cfg.ZoneIDs) > 0 {
		dns.checkZoneIDs()
	} else {
		dns.getZoneIDs()
	}

	dns.Records = make(map[string][]Record)
	dns.getRecords()

	return dns
}

// CheckZoneIDs checks if the zone ids are valid
func (dns *CFDNS) checkZoneIDs() {
	zap.S().Info("Getting zones info")
	reqURL := "https://api.cloudflare.com/client/v4/zones/"

	req, err := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)
	if err != nil {
		zap.S().Fatal(err)
	}

	res, err := dns.HTTPClient.Do(req)
	if err != nil {
		zap.S().Fatal(err)
	}

	type ResponseBody struct {
		Errors   []Error   `json:"errors"`
		Messages []Message `json:"messages"`
		Result   []Zone    `json:"result"`
		Success  bool      `json:"success"`
	}

	var resBody ResponseBody
	err = unmarshalResponse(res.Body, &resBody)
	if err != nil {
		zap.S().Fatal(err)
	}
	res.Body.Close()

	if !resBody.Success || res.StatusCode != http.StatusOK {
		zap.S().Errorf("Error checking zone id, skipping. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
	}

	for _, zoneID := range dns.Cfg.ZoneIDs {
		zap.S().Infof("Checking zone id %s", zoneID)
		isValid := false

		for _, zone := range resBody.Result {
			if zone.ID == zoneID {
				dns.Zones = append(dns.Zones, zone)
				isValid = true
				break
			}
		}

		if !isValid {
			zap.S().Warnf("Zone id %s is not valid, skipping.", zoneID)
		}
	}

	if len(dns.Zones) == 0 {
		zap.S().Fatal("No valid zone ids found")
	}
}

// GetZoneIDs gets the zone id from the zone name
func (dns *CFDNS) getZoneIDs() {
	for _, zoneName := range dns.Cfg.ZoneNames {
		zap.S().Infof("Getting zone id for %s", zoneName)
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", zoneName)

		req, err := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)
		if err != nil {
			zap.S().Fatal(err)
		}

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		type ResponseBody struct {
			Errors   []Error   `json:"errors"`
			Messages []Message `json:"messages"`
			Result   []Zone    `json:"result"`
			Success  bool      `json:"success"`
		}

		var resBody ResponseBody
		err = unmarshalResponse(res.Body, &resBody)
		if err != nil {
			zap.S().Fatal(err)
		}
		res.Body.Close()

		if !resBody.Success || res.StatusCode != http.StatusOK {
			zap.S().Errorf("Error getting zone id, skipping. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
		}

		if len(resBody.Result) == 0 {
			zap.S().Errorf("No zone found with name %s, skipping.", zoneName)
		}

		for _, zone := range resBody.Result {
			if strings.EqualFold(zone.Name, zoneName) {
				dns.Zones = append(dns.Zones, zone)
				break
			}
		}
	}

	if len(dns.Zones) == 0 {
		zap.S().Fatal("No zone ids found")
	}
}

// getRecords gets all the records for the zone
func (dns *CFDNS) getRecords() {
	zap.S().Info("Getting records")
	for _, zone := range dns.Zones {
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zone.ID)

		req, err := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)
		if err != nil {
			zap.S().Fatal(err)
		}

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		type ResponseBody struct {
			Errors   []Error   `json:"errors"`
			Messages []Message `json:"messages"`
			Result   []Record  `json:"result"`
			Success  bool      `json:"success"`
		}

		var resBody ResponseBody
		err = unmarshalResponse(res.Body, &resBody)
		if err != nil {
			zap.S().Fatal(err)
		}
		res.Body.Close()

		if !resBody.Success || res.StatusCode != http.StatusOK {
			zap.S().Fatalf("Error getting records. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
		}

		if len(resBody.Result) == 0 {
			zap.S().Fatalf("No records found for zone id %s", zone.Name)
		}

		recordsMap := make(map[string]Record)
		for _, record := range resBody.Result {
			if record.Type == "A" && (len(dns.Cfg.RecordIDs) == 0 || utils.StringInSlice(record.ID, dns.Cfg.RecordIDs)) {
				recordsMap[record.Name] = record
			}
		}

		if _, ok := dns.Records[zone.Name]; !ok {
			dns.Records[zone.Name] = make([]Record, 0)
		}

		for i, record := range dns.Records[zone.Name] {
			if updatedRecord, ok := recordsMap[record.Name]; ok {
				dns.Records[zone.Name][i] = updatedRecord
				delete(recordsMap, record.Name)
			}
		}

		for _, newRecord := range recordsMap {
			dns.Records[zone.Name] = append(dns.Records[zone.Name], newRecord)
		}

		if len(dns.Records[zone.Name]) == 0 {
			zap.S().Errorf("No records found for zone %s", zone.Name)
		}
	}

	if len(dns.Records) == 0 {
		zap.S().Fatal("No records found")
	}
}

// UpdateRecords updates the records with the current ip
func (dns *CFDNS) UpdateRecords(currentIP string) (updatedRecords map[string][]string, err error) {
	zap.S().Info("Checking records")
	updatedRecords = make(map[string][]string)

	for zoneName, records := range dns.Records {
		for i, record := range records {
			zap.S().Infof("Updating record %s", record.Name)
			reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", record.ZoneID, record.ID)

			payload := strings.NewReader(fmt.Sprintf(`{"content":"%s"}`, currentIP))
			req, err := createCFRequest(http.MethodPatch, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, payload)
			if err != nil {
				zap.S().Fatal(err)
				return updatedRecords, err
			}

			res, err := dns.HTTPClient.Do(req)
			if err != nil {
				zap.S().Fatal(err)
				return updatedRecords, err
			}

			type ResponseBody struct {
				Result   Record    `json:"result"`
				Errors   []Error   `json:"errors"`
				Messages []Message `json:"messages"`
				Success  bool      `json:"success"`
			}

			var resBody ResponseBody
			err = unmarshalResponse(res.Body, &resBody)
			if err != nil {
				zap.S().Fatal(err)
				return updatedRecords, err
			}
			res.Body.Close()
			zap.S().Debugf("Response body: %+v", resBody)

			if !resBody.Success || res.StatusCode != http.StatusOK {
				zap.S().Fatalf("Error updating record. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
				return updatedRecords, err
			}

			records[i] = resBody.Result

			updatedRecords[zoneName] = append(updatedRecords[zoneName], record.Name)
		}
	}

	return updatedRecords, nil
}

// createCFRequest creates an HTTP request with the cloudflare headers
func createCFRequest(method, url, email, authKey string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", authKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// unmarshalResponse unmarshals the response body into the given interface
func unmarshalResponse(reader io.Reader, v interface{}) (err error) {
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return err
	}

	return nil
}

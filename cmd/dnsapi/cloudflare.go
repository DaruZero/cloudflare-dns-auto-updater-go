package dnsapi

import (
	"encoding/json"
	"fmt"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/pkg/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type CFDNS struct {
	HTTPClient HTTPClient
	Cfg        *config.Config
	Records    map[string][]Record
	CurrentIP  string
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

// NewDNS creates a new Dns struct instance
func NewDNS(cfg *config.Config) *CFDNS {
	zap.S().Debug("Creating new Dns struct")
	dns := &CFDNS{
		Cfg: cfg,
	}

	dns.HTTPClient = &http.Client{}

	if len(cfg.ZoneIDs) > 0 {
		dns.CheckZoneIDs()
	} else {
		dns.GetZoneIDs()
	}

	dns.GetCurrentIP()

	dns.Records = make(map[string][]Record)
	dns.GetRecords()

	return dns
}

// CheckZoneIDs checks if the zone ids are valid
func (dns *CFDNS) CheckZoneIDs() {
	zap.S().Info("Getting zones info")
	reqURL := "https://api.cloudflare.com/client/v4/zones/"

	req := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)

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
	unmarshalResponse(res.Body, &resBody)
	res.Body.Close()

	if !resBody.Success || res.StatusCode != 200 {
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
func (dns *CFDNS) GetZoneIDs() {
	for _, zoneName := range dns.Cfg.ZoneNames {
		zap.S().Infof("Getting zone id for %s", zoneName)
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", zoneName)

		req := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)

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
		unmarshalResponse(res.Body, &resBody)
		res.Body.Close()

		if !resBody.Success || res.StatusCode != 200 {
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

// GetCurrentIP gets the current ip address
func (dns *CFDNS) GetCurrentIP() {
	zap.S().Info("Getting current ip")
	reqURL := "https://api.ipify.org"
	timeout := time.Duration(5)

	for {
		req := createRequest(http.MethodGet, reqURL, nil)

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		if res.StatusCode != 200 {
			zap.S().Error("Error getting current ip, retrying in 5 seconds")
			time.Sleep(timeout * time.Second)
			timeout *= 2
			continue
		}

		var bodyString string
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			zap.S().Fatal(err)
		}
		bodyString = string(bodyBytes)
		res.Body.Close()

		zap.S().Info("Successfully got current ip")

		dns.CurrentIP = bodyString
		return
	}
}

// GetRecords gets all the records for the zone
func (dns *CFDNS) GetRecords() {
	zap.S().Info("Getting records")
	for _, zone := range dns.Zones {
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zone.ID)

		req := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)

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
		unmarshalResponse(res.Body, &resBody)
		res.Body.Close()

		if !resBody.Success || res.StatusCode != 200 {
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
func (dns *CFDNS) UpdateRecords() (updatedRecords map[string][]string) {
	zap.S().Info("Checking records")
	updatedRecords = make(map[string][]string)

	for zoneName, records := range dns.Records {
		for _, record := range records {
			if record.Content != dns.CurrentIP {
				zap.S().Infof("Updating record %s", record.Name)
				reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneName, record.ID)

				payload := strings.NewReader(fmt.Sprintf(`{"content":"%s"}`, dns.CurrentIP))
				req := createCFRequest(http.MethodPatch, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, payload)

				res, err := dns.HTTPClient.Do(req)
				if err != nil {
					zap.S().Fatal(err)
				}

				type ResponseBody struct {
					Result   Record    `json:"result"`
					Errors   []Error   `json:"errors"`
					Messages []Message `json:"messages"`
					Success  bool      `json:"success"`
				}

				var resBody ResponseBody
				unmarshalResponse(res.Body, &resBody)
				res.Body.Close()

				if !resBody.Success || res.StatusCode != 200 {
					zap.S().Fatalf("Error updating record. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
				}

				updatedRecords[zoneName] = append(updatedRecords[zoneName], record.Name)
			}
		}
	}

	return updatedRecords
}

// createRequest creates an HTTP request
func createRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		zap.S().Fatal(err)
	}

	return req
}

// createCFRequest creates an HTTP request with the cloudflare headers
func createCFRequest(method, url, email, authKey string, body io.Reader) *http.Request {
	req := createRequest(method, url, body)

	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", authKey)
	req.Header.Set("Content-Type", "application/json")

	return req
}

// unmarshalResponse unmarshals the response body into the given interface
func unmarshalResponse(reader io.Reader, v interface{}) {
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		zap.S().Fatal(err)
	}

	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		zap.S().Fatal(err)
	}
}

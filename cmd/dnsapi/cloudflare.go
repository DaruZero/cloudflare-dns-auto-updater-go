package dnsapi

import (
	"encoding/json"
	"fmt"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/utils"
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
	Cfg        *config.Config
	CurrentIP  string
	ZoneIDs    []string
	Records    map[string][]Record
	HTTPClient HTTPClient
}

type Record struct { //nolint:maligned
	Comment    string `json:"comment"`
	Content    string `json:"content"`
	CreatedOn  string `json:"created_on"`
	Data       Data   `json:"data"`
	ID         string `json:"id"`
	Locked     bool   `json:"locked"`
	Meta       Meta   `json:"meta"`
	ModifiedOn string `json:"modified_on"`
	Name       string `json:"name"`
	Proxiable  bool   `json:"proxiable"`
	Proxied    bool   `json:"proxied"`
	TTL        int    `json:"ttl"`
	Type       string `json:"type"`
	ZoneID     string `json:"zone_id"`
	ZoneName   string `json:"zone_name"`
}

type Meta struct {
	AutoAdded bool   `json:"auto_added"`
	Source    string `json:"source"`
}

type Data struct{}

type ResponseBody struct {
	Success  bool      `json:"success"`
	Errors   []Error   `json:"errors"`
	Messages []Message `json:"messages"`
	Result   []Record  `json:"result"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
	for _, zoneID := range dns.Cfg.ZoneIDs {
		zap.S().Infof("Checking zone id %s", zoneID)
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s", zoneID)

		req := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		var resBody ResponseBody
		unmarshalResponse(res.Body, &resBody)
		res.Body.Close()

		if !resBody.Success || res.StatusCode != 200 {
			zap.S().Errorf("Error checking zone id, skipping. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
		}

		dns.ZoneIDs = append(dns.ZoneIDs, zoneID)
	}

	if len(dns.ZoneIDs) == 0 {
		zap.S().Fatal("No valid zone ids found")
	}
}

// GetZoneIDs gets the zone id from the zone name
func (dns *CFDNS) GetZoneIDs() {
	for _, zoneName := range dns.Cfg.ZoneNames {
		zap.S().Infof("Getting zone id for %s", zoneName)
		zoneID := ""
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", zoneName)

		req := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
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
				zoneID = zone.ID
				break
			}
		}

		dns.ZoneIDs = append(dns.ZoneIDs, zoneID)
	}

	if len(dns.ZoneIDs) == 0 {
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
	for _, zoneID := range dns.ZoneIDs {
		reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)

		req := createCFRequest(http.MethodGet, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, nil)

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		var resBody ResponseBody
		unmarshalResponse(res.Body, &resBody)
		res.Body.Close()

		if !resBody.Success || res.StatusCode != 200 {
			zap.S().Fatalf("Error getting records. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
		}

		if len(resBody.Result) == 0 {
			zap.S().Fatalf("No records found for zone id %s", zoneID)
		}

		if len(dns.Cfg.RecordIDs) > 0 {
			for _, record := range resBody.Result {
				if record.Type == "A" && utils.StringInSlice(record.ID, dns.Cfg.RecordIDs) {
					zap.S().Info("Found record for given id")
					dns.Records[zoneID] = append(dns.Records[zoneID], record)
				}
			}
			zap.S().Errorf("Record id not found, skipping.")
		}

		for _, record := range resBody.Result {
			if record.Type == "A" {
				dns.Records[zoneID] = append(dns.Records[zoneID], record)
			}
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

				payload := strings.NewReader(fmt.Sprintf(`{"content":"%s","name":"%s","ttl":"%d"}`, dns.CurrentIP, record.Name, record.TTL))
				req := createCFRequest(http.MethodPut, reqURL, dns.Cfg.Email, dns.Cfg.AuthKey, payload)

				res, err := dns.HTTPClient.Do(req)
				if err != nil {
					zap.S().Fatal(err)
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

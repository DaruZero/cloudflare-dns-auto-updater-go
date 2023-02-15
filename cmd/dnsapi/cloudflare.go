package dnsapi

import (
	"cloudflare-dns-auto-updater-go/cmd/config"
	"encoding/json"
	"fmt"
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
	ZoneID     string
	Records    []Record
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

	if cfg.ZoneID != "" {
		dns.ZoneID = cfg.ZoneID
	} else {
		dns.ZoneID = dns.GetZoneID(cfg.ZoneName)
	}

	dns.CurrentIP = dns.GetCurrentIP()

	dns.Records = dns.GetRecords()

	return dns
}

// GetZoneID gets the zone id from the zone name
func (dns *CFDNS) GetZoneID(zoneName string) string {
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
		zap.S().Fatalf("Error getting zone id. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
	}

	if len(resBody.Result) == 0 {
		zap.S().Fatalf("No zone found with name %s", zoneName)
	}

	zoneID = resBody.Result[0].ID

	zap.S().Info("Successfully got zone id")

	return zoneID
}

// GetCurrentIP gets the current ip address
func (dns *CFDNS) GetCurrentIP() string {
	zap.S().Info("Getting current ip")
	reqURL := "https://api.ipify.org"

	for {
		req := createRequest(http.MethodGet, reqURL, nil)

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		if res.StatusCode != 200 {
			zap.S().Error("Error getting current ip, retrying in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}

		var bodyString string
		unmarshalResponse(res.Body, &bodyString)
		res.Body.Close()

		zap.S().Info("Successfully got current ip")

		return bodyString
	}
}

// GetRecords gets all the records for the zone
func (dns *CFDNS) GetRecords() []Record {
	zap.S().Info("Getting records")
	reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", dns.ZoneID)

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

	if dns.Cfg.RecordID != "" {
		for _, record := range resBody.Result {
			if record.ID == dns.Cfg.RecordID {
				zap.S().Info("Found record for given id")
				return []Record{record}
			}
		}
		zap.S().Fatal("Record id not found")
	}

	var records []Record
	for _, record := range resBody.Result {
		if record.Type == "A" || record.Type == "AAAA" {
			records = append(records, record)
		}
	}

	if len(records) == 0 {
		zap.S().Fatal("No A records found")
	}

	zap.S().Info("Successfully got records")

	return records
}

// UpdateRecords updates the records with the current ip
func (dns *CFDNS) UpdateRecords() (updatedRecords []string, updated bool) {
	zap.S().Info("Checking records")

	updated = false

	for _, record := range dns.Records {
		if record.Content != dns.CurrentIP {
			zap.S().Infof("Updating record %s", record.Name)
			updated = true
			reqURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", dns.ZoneID, record.ID)

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
				zap.S().Fatalf("Error updating records. HTTP status code: %d. Response body: %v", res.StatusCode, resBody)
			}

			updatedRecords = append(updatedRecords, record.Name)

			zap.S().Infof("Updated record %s", record.Name)
		}
	}

	return
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

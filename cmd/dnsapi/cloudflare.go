package dnsapi

import (
	"cloudflare-dns-auto-updater-go/cmd/config"
	"cloudflare-dns-auto-updater-go/cmd/utils"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type CFDNS struct {
	Cfg        *config.Config
	CurrentIP  string
	ZoneID     string
	Records    []Record
	HTTPClient *http.Client
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

	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones", nil)
	if err != nil {
		zap.S().Fatal(err)
	}

	req.Header.Set("X-Auth-Email", dns.Cfg.Email)
	req.Header.Set("X-Auth-Key", dns.Cfg.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	url := utils.SanitizeString(req.URL.String())
	zap.S().Debugf("Sending request to %s", url)

	res, err := dns.HTTPClient.Do(req)
	if err != nil {
		zap.S().Fatal(err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		zap.S().Fatal(err)
	}

	var resBody ResponseBody
	err = json.Unmarshal(bodyBytes, &resBody)
	if err != nil {
		zap.S().Fatal(err)
	}

	if !resBody.Success || res.StatusCode != 200 {
		zap.S().Fatalf("Error getting zone id. HTTP status code: %d. Response body: %s", res.StatusCode, string(bodyBytes))
	}

	for _, z := range resBody.Result {
		if z.Name == zoneName {
			zoneID = z.ID
		}
	}

	zap.S().Info("Successfully got zone id")
	res.Body.Close()

	return zoneID
}

// GetCurrentIP gets the current ip address
func (dns *CFDNS) GetCurrentIP() string {
	zap.S().Info("Getting current ip")

	for {
		req, err := http.NewRequest("GET", "https://api.ipify.org", nil)
		if err != nil {
			zap.S().Fatal(err)
		}

		url := utils.SanitizeString(req.URL.String())
		zap.S().Debugf("Sending request to %s", url)

		res, err := dns.HTTPClient.Do(req)
		if err != nil {
			zap.S().Fatal(err)
		}

		if res.StatusCode != 200 {
			zap.S().Error("Error getting current ip, retrying in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			zap.S().Fatal(err)
		}

		zap.S().Info("Successfully got current ip")
		res.Body.Close()

		return string(bodyBytes)
	}
}

// GetRecords gets all the records for the zone
func (dns *CFDNS) GetRecords() []Record {
	zap.S().Info("Getting records")

	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones/"+dns.ZoneID+"/dns_records", nil)
	if err != nil {
		zap.S().Fatal(err)
	}

	req.Header.Set("X-Auth-Email", dns.Cfg.Email)
	req.Header.Set("X-Auth-Key", dns.Cfg.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	url := utils.SanitizeString(req.URL.String())
	zap.S().Debugf("Sending request to %s", url)

	res, err := dns.HTTPClient.Do(req)
	if err != nil {
		zap.S().Fatal(err)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		zap.S().Fatal(err)
	}

	var resBody ResponseBody
	err = json.Unmarshal(bodyBytes, &resBody)
	if err != nil {
		zap.S().Fatal(err)
	}

	if !resBody.Success || res.StatusCode != 200 {
		zap.S().Fatalf("Error getting records. HTTP status code: %d. Response body: %s", res.StatusCode, string(bodyBytes))
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
	res.Body.Close()

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

			payload := strings.NewReader(fmt.Sprintf(`{"content":"%s","name":"%s","ttl":"%d"}`, dns.CurrentIP, record.Name, record.TTL))
			req, err := http.NewRequest("PUT", "https://api.cloudflare.com/client/v4/zones/"+dns.ZoneID+"/dns_records/"+record.ID, payload)
			if err != nil {
				zap.S().Fatal(err)
			}

			req.Header.Set("X-Auth-Email", dns.Cfg.Email)
			req.Header.Set("X-Auth-Key", dns.Cfg.AuthKey)
			req.Header.Set("Content-Type", "application/json")

			url := utils.SanitizeString(req.URL.String())
			zap.S().Debugf("Sending request to %s", url)

			res, err := dns.HTTPClient.Do(req)
			if err != nil {
				zap.S().Fatal(err)
			}

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				zap.S().Fatal(err)
			}

			var resBody ResponseBody
			err = json.Unmarshal(bodyBytes, &resBody)
			if err != nil {
				zap.S().Fatal(err)
			}

			if !resBody.Success || res.StatusCode != 200 {
				zap.S().Fatalf("Error updating records. HTTP status code: %d. Response body: %s", res.StatusCode, string(bodyBytes))
			}

			updatedRecords = append(updatedRecords, record.Name)

			zap.S().Infof("Updated record %s", record.Name)
			res.Body.Close()
		}
	}

	return
}

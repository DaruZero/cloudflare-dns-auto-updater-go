package main

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type Dns struct {
	Cfg       *Config
	CurrentIp string
	ZoneId    string
	Records   []Record
}

type Record struct {
	Comment   string   `json:"comment"`
	Content   string   `json:"content"`
	CreatedOn string   `json:"created_on"`
	Data      struct{} `json:"data"`
	Id        string   `json:"id"`
	Locked    bool     `json:"locked"`
	Meta      struct {
		AutoAdded bool   `json:"auto_added"`
		Source    string `json:"source"`
	}
	ModifiedOn string `json:"modified_on"`
	Name       string `json:"name"`
	Proxiable  bool   `json:"proxiable"`
	Proxied    bool   `json:"proxied"`
	Ttl        int    `json:"ttl"`
	Type       string `json:"type"`
	ZoneId     string `json:"zone_id"`
	ZoneName   string `json:"zone_name"`
}

// NewDns creates a new Dns struct instance
func NewDns(cfg *Config) *Dns {
	zap.S().Info("Creating new Dns struct")
	dns := &Dns{
		Cfg: cfg,
	}

	if cfg.ZoneId != "" {
		dns.ZoneId = cfg.ZoneId
	} else {
		dns.ZoneId = dns.GetZoneId(cfg.ZoneName)
	}

	dns.CurrentIp = dns.GetCurrentIp()

	dns.Records = dns.GetRecords()

	return dns
}

// GetZoneId gets the zone id from the zone name
func (dns *Dns) GetZoneId(zoneName string) string {
	zap.S().Infof("Getting zone id for %s", zoneName)
	zoneId := ""
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones", nil)
	if err != nil {
		zap.S().Fatal(err)
	}

	req.Header.Set("X-Auth-Email", dns.Cfg.Email)
	req.Header.Set("X-Auth-Key", dns.Cfg.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	zap.S().Debugf("Sending request to %s", req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		zap.S().Fatal(err)
	}

	if res.StatusCode != 200 {
		zap.S().Fatal("Error getting zone id. HTTP status code: ", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		zap.S().Fatal(err)
	}

	var resBody struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Messages []string `json:"messages"`
		Result   []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}

	err = json.Unmarshal(bodyBytes, &resBody)
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Debugf("Response body: %+v", resBody)

	for _, z := range resBody.Result {
		if z.Name == zoneName {
			zoneId = z.Id
		}
	}

	zap.S().Info("Successfully got zone id")

	return zoneId
}

// GetCurrentIp gets the current ip address
func (dns *Dns) GetCurrentIp() string {
	zap.S().Info("Getting current ip")
	for {
		client := http.Client{}

		req, err := http.NewRequest("GET", "http://api.ipify.org", nil)
		if err != nil {
			zap.S().Fatal(err)
		}

		zap.S().Debugf("Sending request to %s", req.URL.String())

		res, err := client.Do(req)
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

		zap.S().Debugf("Response body: %s", string(bodyBytes))
		zap.S().Info("Successfully got current ip")

		return string(bodyBytes)
	}
}

// GetRecords gets all the records for the zone
func (dns *Dns) GetRecords() []Record {
	zap.S().Info("Getting records")
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones/"+dns.ZoneId+"/dns_records", nil)
	if err != nil {
		zap.S().Fatal(err)
	}

	req.Header.Set("X-Auth-Email", dns.Cfg.Email)
	req.Header.Set("X-Auth-Key", dns.Cfg.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	zap.S().Debugf("Sending request to %s", req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		zap.S().Fatal(err)
	}

	if res.StatusCode != 200 {
		zap.S().Fatal("Error getting records")
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		zap.S().Fatal(err)
	}

	var resBody struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Messages []string `json:"messages"`
		Result   []Record `json:"result"`
	}

	err = json.Unmarshal(bodyBytes, &resBody)
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Debugf("Response body: %+v", resBody)

	if dns.Cfg.RecordId != "" {
		for _, record := range resBody.Result {
			if record.Id == dns.Cfg.RecordId {
				zap.S().Info("Found record for given id")
				return []Record{record}
			}
		}
		zap.S().Fatal("Record id not found")
	}

	zap.S().Info("Successfully got records")

	return resBody.Result
}

// UpdateRecords updates the records with the current ip
func (dns *Dns) UpdateRecords() (updatedRecords []string, updated bool) {
	zap.S().Info("Checking records")

	updated = false

	for _, record := range dns.Records {
		if record.Content != dns.CurrentIp {
			zap.S().Infof("Updating record %s", record.Name)
			updated = true
			client := http.Client{}

			payload := strings.NewReader(`{"content":"` + dns.CurrentIp)
			req, err := http.NewRequest("PUT", "https://api.cloudflare.com/client/v4/zones/"+dns.ZoneId+"/dns_records/"+record.Id, payload)
			if err != nil {
				zap.S().Fatal(err)
			}

			req.Header.Set("X-Auth-Email", dns.Cfg.Email)
			req.Header.Set("X-Auth-Key", dns.Cfg.AuthKey)
			req.Header.Set("Content-Type", "application/json")

			zap.S().Debugf("Sending request to %s", req.URL.String())

			res, err := client.Do(req)
			if err != nil {
				zap.S().Fatal(err)
			}

			if res.StatusCode != 200 {
				zap.S().Fatal("Error updating records")
			}

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				zap.S().Fatal(err)
			}

			var resBody struct {
				Success bool `json:"success"`
				Errors  []struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"errors"`
				Messages []string `json:"messages"`
				Result   Record   `json:"result"`
			}

			err = json.Unmarshal(bodyBytes, &resBody)
			if err != nil {
				zap.S().Fatal(err)
			}

			updatedRecords = append(updatedRecords, record.Name)

			zap.S().Debugf("Response body: %+v", resBody)

			zap.S().Info("Updated record ", record.Name, " to ", resBody.Result.Content)
		}
	}

	return
}

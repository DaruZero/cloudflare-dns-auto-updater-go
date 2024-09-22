package dnsapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the API client being tested.
	client *cloudflare.API

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

func init() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// disable rate limits and retries in testing - prepended so any provided value overrides this
	opts := []cloudflare.Option{cloudflare.UsingRateLimit(100000), cloudflare.UsingRetryPolicy(0, 0, 0)}

	// Cloudflare client configured to use test server
	client, _ = cloudflare.New("deadbeef", "cloudflare@example.org", opts...)
	client.BaseURL = server.URL
}

func teardown() {
	server.Close()
}

func parsePage(t *testing.T, total int, s string) (int, bool) {
	if s == "" {
		return 1, true
	}

	page, err := strconv.Atoi(s)
	if !assert.NoError(t, err) {
		return 0, false
	}

	if !assert.LessOrEqual(t, page, total) || !assert.GreaterOrEqual(t, page, 1) {
		return 0, false
	}

	return page, true
}

func TestDns_getRecords(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case !assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method):
			return
		case !assert.Equal(t, "50", r.URL.Query().Get("per_page")):
			return
		}

		page, ok := parsePage(t, totalPage, r.URL.Query().Get("page"))
		if !ok {
			return
		}

		start := (page - 1) * 50

		count := 50
		if page == totalPage {
			count = total - start
		}

		w.Header().Set("content-type", "application/json")
		err := json.NewEncoder(w).Encode(mockZonesResponse(total, page, start, count))
		assert.NoError(t, err)
	}

	mux.HandleFunc("/zones", handler)

	tests := []struct {
		name                string
		mockResponse        string
		recordIDs           []string
		expectedRecordIDs   []string
		expectedNames       []string
		expectedTypes       []string
		expectedContents    []string
		expectedRecordsMaps int
		wantErr             bool
	}{
		{
			name:                "NoRecordID",
			recordIDs:           nil,
			expectedRecordsMaps: 1,
			expectedRecordIDs:   []string{"testRecordID"},
			expectedNames:       []string{"testRecordName"},
			expectedTypes:       []string{"A"},
			expectedContents:    []string{"testContent"},
			wantErr:             false,
		},
		{
			name:                "WithRecordID",
			recordIDs:           []string{"testRecordID"},
			expectedRecordsMaps: 1,
			expectedRecordIDs:   []string{"testRecordID"},
			expectedNames:       []string{"testRecordName"},
			expectedTypes:       []string{"A"},
			expectedContents:    []string{"testContent"},
			wantErr:             false,
		},
		{
			name:                "InvalidRecordID",
			recordIDs:           []string{"testRecordID"},
			expectedRecordsMaps: 0,
			expectedRecordIDs:   []string{},
			expectedNames:       []string{},
			expectedTypes:       []string{},
			expectedContents:    []string{},
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AuthKey:   "deadbeef",
				Email:     "test@example.org",
				RecordIDs: tt.recordIDs,
				ZoneIDs:   []string{"testZoneID"},
			}

			dns, err := New(cfg)
			if err != nil {
				t.Fatal(err)
			}

			err = dns.getRecords()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRecords() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(dns.Records) != tt.expectedRecordsMaps {
				t.Fatalf("GetRecords() = %d; want 1", len(dns.Records))
			}

			zoneName := dns.Zones[0].Name

			if len(dns.Records[zoneName]) != len(tt.expectedRecordIDs) {
				t.Fatalf("GetRecords() = %d; want %d", len(dns.Records[zoneName]), len(tt.expectedRecordIDs))
			}

			for i, record := range dns.Records[zoneName] {
				if record.ID != tt.expectedRecordIDs[i] {
					t.Errorf("GetRecords() = %s; want %s", record.ID, tt.expectedRecordIDs[i])
				}

				if record.Name != tt.expectedNames[i] {
					t.Errorf("GetRecords() = %s; want %s", record.Name, tt.expectedNames[i])
				}

				if record.Type != tt.expectedTypes[i] {
					t.Errorf("GetRecords() = %s; want %s", record.Type, tt.expectedTypes[i])
				}

				if record.Content != tt.expectedContents[i] {
					t.Errorf("GetRecords() = %s; want %s", record.Content, tt.expectedContents[i])
				}
			}
		})
	}
}

// func TestDns_UpdateRecord(t *testing.T) {
// 	tests := []struct {
// 		initialRecords  map[string][]cloudflare.DNSRecord
// 		expectedRecords map[string][]string
// 		updatedRecords  map[string][]cloudflare.DNSRecord
// 		name            string
// 		updateIP        string
// 		mockResponse    string
// 		wantErr         bool
// 	}{
// 		{
// 			name: "UpdateRecordSuccess",
// 			initialRecords: map[string][]cloudflare.DNSRecord{
// 				"testZoneID": {
// 					{
// 						ID:      "testRecordID",
// 						Name:    "testRecordName",
// 						Type:    "A",
// 						Content: "testIPOld",
// 					},
// 				},
// 			},
// 			updateIP:     "testIPNew",
// 			mockResponse: `{"success":true,"errors":[],"messages":[],"result":{"id":"testRecordID", "name": "testRecordName", "type": "A", "content": "testIPNew"}}`,
// 			expectedRecords: map[string][]string{
// 				"testZoneID": {"testRecordName"},
// 			},
// 			updatedRecords: map[string][]cloudflare.DNSRecord{
// 				"testZoneID": {
// 					{
// 						ID:      "testRecordID",
// 						Name:    "testRecordName",
// 						Type:    "A",
// 						Content: "testIPNew",
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "UpdateRecordFail",
// 			initialRecords: map[string][]cloudflare.DNSRecord{
// 				"testZoneID": {
// 					{
// 						ID:      "testRecordID",
// 						Name:    "testRecordName",
// 						Type:    "A",
// 						Content: "testIPOld",
// 					},
// 				},
// 			},
// 			updateIP:        "testIPNew",
// 			mockResponse:    `{"success":false,"errors":[{"code":1004,"message":"DNS Validation Error","error_chain":[{"code":9003,"message":"Invalid IP","error_chain":[]}]}],"messages":[],"result":null}`,
// 			expectedRecords: map[string][]string{},
// 			updatedRecords:  map[string][]cloudflare.DNSRecord{},
// 			wantErr:         true,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cfg := &config.Config{
// 				AuthKey: "deadbeef",
// 				Email:   "test@example.org",
// 			}
//
// 			dns := &CFDNS{
// 				Cfg:     cfg,
// 				Records: tt.initialRecords,
// 			}
//
// 			updatedRecords, err := dns.UpdateRecords(tt.updateIP)
// 			if (err != nil) != tt.wantErr {
// 				t.Fatalf("UpdateRecords() error = %v, wantErr %v", err, tt.wantErr)
// 			}
//
// 			for zoneID, updatedZoneRecords := range tt.expectedRecords {
// 				if len(updatedRecords[zoneID]) != len(updatedZoneRecords) {
// 					t.Fatalf("UpdateRecords() = %d; want %d", len(updatedRecords[zoneID]), len(updatedZoneRecords))
// 				}
//
// 				for i, expectedRecordName := range updatedZoneRecords {
// 					if updatedRecords[zoneID][i] != expectedRecordName {
// 						t.Errorf("UpdateRecords() = %s; want %s", updatedRecords[zoneID][i], expectedRecordName)
// 					}
// 				}
// 			}
//
// 			for zoneID, updatedZoneRecords := range tt.updatedRecords {
// 				if len(dns.Records[zoneID]) != len(updatedZoneRecords) {
// 					t.Fatalf("UpdateRecords() = %d; want %d", len(dns.Records[zoneID]), len(updatedZoneRecords))
// 				}
//
// 				for i, expectedRecord := range updatedZoneRecords {
// 					if dns.Records[zoneID][i].Content != expectedRecord.Content {
// 						t.Errorf("UpdateRecords() = %v; want %v", dns.Records[zoneID][i], expectedRecord)
// 					}
// 				}
// 			}
// 		})
// 	}
// }

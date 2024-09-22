package dnsapi

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/daruzero/cloudflare-dns-auto-updater-go/internal/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/test/mocks"
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func TestDns_checkZoneIDs(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		expectedZoneIDs []string
		expectedNames   []string
		expectedZones   int
		wantErr         bool
	}{
		{
			name:            "ValidZoneIDs",
			mockResponse:    `{"success":true,"errors":[],"messages":[],"result":[{"id":"testZoneID1", "name": "testZoneName1"}, {"id":"testZoneID2", "name": "testZoneName2"}, {"id":"testZoneID3", "name": "testZoneName"}]}`,
			expectedZones:   2,
			expectedZoneIDs: []string{"testZoneID1", "testZoneID2"},
			expectedNames:   []string{"testZoneName1", "testZoneName2"},
			wantErr:         false,
		},
		{
			name:            "InvalidZoneIDs",
			mockResponse:    `{"success":true,"errors":[],"messages":[],"result":[{"id":"testZoneID1", "name": "testZoneName1"}, {"id":"testZoneID4", "name": "testZoneName4"}, {"id":"testZoneID5", "name": "testZoneName5"}]}`,
			expectedZones:   1,
			expectedZoneIDs: []string{"testZoneID1"},
			expectedNames:   []string{"testZoneName1"},
			wantErr:         false,
		},
		{
			name:            "NoZoneIDs",
			mockResponse:    `{"success":true,"errors":[],"messages":[],"result":[{"id":"testZoneID3", "name": "testZoneName3"}]}`,
			expectedZones:   0,
			expectedZoneIDs: []string{},
			expectedNames:   []string{},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := io.NopCloser(bytes.NewReader([]byte(tt.mockResponse)))
					return &http.Response{
						StatusCode: 200,
						Body:       body,
					}, nil
				},
			}

			cfg := &config.Config{
				AuthKey: "testAuthKey",
				Email:   "testEmail",
				ZoneIDs: []string{"testZoneID1", "testZoneID2"},
			}

			dns := &CFDNS{
				Cfg:        cfg,
				HTTPClient: mockClient,
			}

			err := dns.checkZoneIDs()
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckZoneIDs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(dns.Zones) != tt.expectedZones {
				t.Fatalf("CheckZoneIDs() = %d; want %d", len(dns.Zones), tt.expectedZones)
			}

			for i, zone := range dns.Zones {
				if zone.ID != tt.expectedZoneIDs[i] {
					t.Errorf("CheckZoneIDs() = %s; want %s", zone.ID, tt.expectedZoneIDs[i])
				}

				if zone.Name != tt.expectedNames[i] {
					t.Errorf("CheckZoneIDs() = %s; want %s", zone.Name, tt.expectedNames[i])
				}
			}
		})
	}
}

func TestDns_getZoneIDs(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		expectedZoneIDs []string
		expectedNames   []string
		expectedZones   int
		wantErr         bool
	}{
		{
			name:            "ValidZoneNames",
			mockResponse:    `{"success":true,"errors":[],"messages":[],"result":[{"id":"testID", "name": "testZoneName"}, {"id":"testID2", "name": "testZoneName2"}]}`,
			expectedZones:   2,
			expectedZoneIDs: []string{"testID", "testID2"},
			expectedNames:   []string{"testZoneName", "testZoneName2"},
			wantErr:         false,
		},
		{
			name:            "InvalidZoneNames",
			mockResponse:    `{"success":true,"errors":[],"messages":[],"result":[]}`,
			expectedZones:   0,
			expectedZoneIDs: []string{},
			expectedNames:   []string{},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := io.NopCloser(bytes.NewReader([]byte(tt.mockResponse)))
					return &http.Response{
						StatusCode: 200,
						Body:       body,
					}, nil
				},
			}

			cfg := &config.Config{
				AuthKey:   "testAuthKey",
				Email:     "testEmail",
				ZoneNames: []string{"testZoneName", "testZoneName2"},
			}

			dns := &CFDNS{
				Cfg:        cfg,
				HTTPClient: mockClient,
			}

			err := dns.getZoneIDs()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetZoneIDs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(dns.Zones) != tt.expectedZones {
				t.Fatalf("GetZoneIDs() = %d; want %d", len(dns.Zones), tt.expectedZones)
			}

			for i, zone := range dns.Zones {
				if zone.ID != tt.expectedZoneIDs[i] {
					t.Errorf("GetZoneIDs() = %s; want %s", zone.ID, tt.expectedZoneIDs[i])
				}

				if zone.Name != tt.expectedNames[i] {
					t.Errorf("GetZoneIDs() = %s; want %s", zone.Name, tt.expectedNames[i])
				}
			}
		})
	}
}

func TestDns_getRecords(t *testing.T) {
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
			mockResponse:        `{"success":true,"errors":[],"messages":[],"result":[{"id":"testRecordID", "name": "testRecordName", "type": "A", "content": "testContent"}]}`,
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
			mockResponse:        `{"success":true,"errors":[],"messages":[],"result":[{"id":"testRecordID", "name": "testRecordName", "type": "A", "content": "testContent"}]}`,
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
			mockResponse:        `{"success":true,"errors":[],"messages":[],"result":[{"id":"testRecordID2", "name": "testRecordName2", "type": "A", "content": "testContent2"}]}`,
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
			mockClient := &mocks.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := io.NopCloser(bytes.NewReader([]byte(tt.mockResponse)))
					return &http.Response{
						StatusCode: 200,
						Body:       body,
					}, nil
				},
			}

			cfg := &config.Config{
				AuthKey:   "testAuthKey",
				Email:     "testEmail",
				RecordIDs: tt.recordIDs,
			}

			dns := &CFDNS{
				Cfg:        cfg,
				HTTPClient: mockClient,
				Records:    make(map[string][]Record),
				Zones:      []Zone{{ID: "testZoneID", Name: "testZoneName"}},
			}

			err := dns.getRecords()
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

func TestDns_UpdateRecord(t *testing.T) {
	tests := []struct {
		initialRecords  map[string][]Record
		expectedRecords map[string][]string
		updatedRecords  map[string][]Record
		name            string
		updateIP        string
		mockResponse    string
		wantErr         bool
	}{
		{
			name: "UpdateRecordSuccess",
			initialRecords: map[string][]Record{
				"testZoneID": {
					{
						ID:      "testRecordID",
						Name:    "testRecordName",
						Type:    "A",
						Content: "testIPOld",
					},
				},
			},
			updateIP:     "testIPNew",
			mockResponse: `{"success":true,"errors":[],"messages":[],"result":{"id":"testRecordID", "name": "testRecordName", "type": "A", "content": "testIPNew"}}`,
			expectedRecords: map[string][]string{
				"testZoneID": {"testRecordName"},
			},
			updatedRecords: map[string][]Record{
				"testZoneID": {
					{
						ID:      "testRecordID",
						Name:    "testRecordName",
						Type:    "A",
						Content: "testIPNew",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "UpdateRecordFail",
			initialRecords: map[string][]Record{
				"testZoneID": {
					{
						ID:      "testRecordID",
						Name:    "testRecordName",
						Type:    "A",
						Content: "testIPOld",
					},
				},
			},
			updateIP:        "testIPNew",
			mockResponse:    `{"success":false,"errors":[{"code":1004,"message":"DNS Validation Error","error_chain":[{"code":9003,"message":"Invalid IP","error_chain":[]}]}],"messages":[],"result":null}`,
			expectedRecords: map[string][]string{},
			updatedRecords:  map[string][]Record{},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := io.NopCloser(bytes.NewReader([]byte(tt.mockResponse)))
					return &http.Response{
						StatusCode: 200,
						Body:       body,
					}, nil
				},
			}

			cfg := &config.Config{
				AuthKey: "testAuthKey",
				Email:   "testEmail",
			}

			dns := &CFDNS{
				Cfg:        cfg,
				HTTPClient: mockClient,
				Records:    tt.initialRecords,
			}

			updatedRecords, err := dns.UpdateRecords(tt.updateIP)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateRecords() error = %v, wantErr %v", err, tt.wantErr)
			}

			for zoneID, updatedZoneRecords := range tt.expectedRecords {
				if len(updatedRecords[zoneID]) != len(updatedZoneRecords) {
					t.Fatalf("UpdateRecords() = %d; want %d", len(updatedRecords[zoneID]), len(updatedZoneRecords))
				}

				for i, expectedRecordName := range updatedZoneRecords {
					if updatedRecords[zoneID][i] != expectedRecordName {
						t.Errorf("UpdateRecords() = %s; want %s", updatedRecords[zoneID][i], expectedRecordName)
					}
				}
			}

			for zoneID, updatedZoneRecords := range tt.updatedRecords {
				if len(dns.Records[zoneID]) != len(updatedZoneRecords) {
					t.Fatalf("UpdateRecords() = %d; want %d", len(dns.Records[zoneID]), len(updatedZoneRecords))
				}

				for i, expectedRecord := range updatedZoneRecords {
					if dns.Records[zoneID][i] != expectedRecord {
						t.Errorf("UpdateRecords() = %v; want %v", dns.Records[zoneID][i], expectedRecord)
					}
				}
			}
		})
	}
}

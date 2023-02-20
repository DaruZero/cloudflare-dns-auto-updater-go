package dnsapi

import (
	"bytes"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/config"
	"github.com/daruzero/cloudflare-dns-auto-updater-go/cmd/mocks"
	"io"
	"net/http"
	"testing"
)

// TestDns_GetZoneIDs tests the GetZoneId method
func TestDns_GetZoneIDs(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey:   "testAuthKey",
		Email:     "testEmail",
		ZoneNames: []string{"testZoneName"},
	}
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			json := `{"success":true,"errors":[],"messages":[],"result":[{"id":"testID", "name": "testZoneName"}]}`
			body := io.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{
				StatusCode: 200,
				Body:       body,
			}, nil
		},
	}
	// Create a new dns object
	dns := &CFDNS{
		Cfg:        cfg,
		HTTPClient: mockClient,
	}
	// Get the zone id
	dns.GetZoneIDs()
	// Check the zone id
	if len(dns.ZoneIDs) != 1 {
		t.Fatalf("GetZoneIDs() = %d; want 1", len(dns.ZoneIDs))
	}
	if dns.ZoneIDs[0] != "testID" {
		t.Errorf("GetZoneIDs() = %s; want testID", dns.ZoneIDs[0])
	}
}

// TestDns_GetCurrentIP tests the GetCurrentIP method
func TestDns_GetCurrentIP(t *testing.T) {
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`testIP`))),
			}, nil
		},
	}
	// Create a new dns object
	dns := &CFDNS{
		HTTPClient: mockClient,
	}
	// Get the current IP
	dns.GetCurrentIP()
	// Check the current IP
	if dns.CurrentIP != "testIP" {
		t.Errorf("GetCurrentIP() = %s; want testIP", dns.CurrentIP)
	}
}

// TestDns_GetRecordsNoRecordID tests the GetRecords method
func TestDns_GetRecordsNoRecordID(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey: "testAuthKey",
		Email:   "testEmail",
	}
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			json := `{"success":true,"errors":[],"messages":[],"result":[{"id":"testRecordID", "name": "testRecordName", "type": "A", "content": "testContent"}]}`
			body := io.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{
				StatusCode: 200,
				Body:       body,
			}, nil
		},
	}
	// Create a new dns object
	dns := &CFDNS{
		Cfg:        cfg,
		HTTPClient: mockClient,
		Records:    make(map[string][]Record),
		ZoneIDs:    []string{"testZoneID"},
	}
	// Get the records
	dns.GetRecords()
	// Check the records
	if len(dns.Records) != 1 {
		t.Fatalf("GetRecords() = %d; want 1", len(dns.Records))
	}
	if len(dns.Records["testZoneID"]) != 1 {
		t.Logf("%+v", dns.Records)
		t.Logf("%+v", dns.Records["testZoneID"])
		t.Fatalf("GetRecords() = %d; want 1", len(dns.Records["testZoneID"]))
	}
	if dns.Records["testZoneID"][0].ID != "testRecordID" {
		t.Errorf("GetRecords() = %s; want testRecordID", dns.Records["testZoneID"][0].ID)
	}
	if dns.Records["testZoneID"][0].Name != "testRecordName" {
		t.Errorf("GetRecords() = %s; want testRecordName", dns.Records["testZoneID"][0].Name)
	}
	if dns.Records["testZoneID"][0].Type != "A" {
		t.Errorf("GetRecords() = %s; want A", dns.Records["testZoneID"][0].Type)
	}
	if dns.Records["testZoneID"][0].Content != "testContent" {
		t.Errorf("GetRecords() = %s; want testContent", dns.Records["testZoneID"][0].Content)
	}
}

// TestDns_GetRecordsWithRecordID tests the GetRecords method
func TestDns_GetRecordsWithRecordID(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey:   "testAuthKey",
		Email:     "testEmail",
		RecordIDs: []string{"testRecordID"},
	}
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			json := `{"success":true,"errors":[],"messages":[],"result":[{"id":"testRecordID", "name": "testRecordName", "type": "A", "content": "testContent"}]}`
			body := io.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{
				StatusCode: 200,
				Body:       body,
			}, nil
		},
	}
	// Create a new dns object
	dns := &CFDNS{
		Cfg:        cfg,
		HTTPClient: mockClient,
		Records:    make(map[string][]Record),
		ZoneIDs:    []string{"testZoneID"},
	}
	// Get the records
	dns.GetRecords()
	// Check the records
	if len(dns.Records) != 1 {
		t.Fatalf("GetRecords() = %d; want 1", len(dns.Records))
	}
	if len(dns.Records["testZoneID"]) != 1 {
		t.Fatalf("GetRecords() = %d; want 1", len(dns.Records["testZoneID"]))
	}
	if dns.Records["testZoneID"][0].ID != "testRecordID" {
		t.Fatalf("GetRecords() = %s; want testRecordID", dns.Records["testZoneID"][0].ID)
	}
	if dns.Records["testZoneID"][0].Name != "testRecordName" {
		t.Errorf("GetRecords() = %s; want testRecordName", dns.Records["testZoneID"][0].Name)
	}
	if dns.Records["testZoneID"][0].Type != "A" {
		t.Errorf("GetRecords() = %s; want A", dns.Records["testZoneID"][0].Type)
	}
	if dns.Records["testZoneID"][0].Content != "testContent" {
		t.Errorf("GetRecords() = %s; want testContent", dns.Records["testZoneID"][0].Content)
	}
}

// TestDns_UpdateRecord tests the UpdateRecord method
func TestDns_UpdateRecord(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey: "testAuthKey",
		Email:   "testEmail",
	}
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			json := `{"success":true,"errors":[],"messages":[],"result":[]}`
			body := io.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{
				StatusCode: 200,
				Body:       body,
			}, nil
		},
	}
	// Create a new dns object
	dns := &CFDNS{
		Cfg:        cfg,
		HTTPClient: mockClient,
		CurrentIP:  "testIPNew",
		Records: map[string][]Record{
			"testZoneID": {
				{
					ID:      "testRecordID",
					Name:    "testRecordName",
					Type:    "A",
					Content: "testIPOld",
				},
			},
		},
	}
	// Update the records
	updatedRecords := dns.UpdateRecords()
	// Check the records
	if len(updatedRecords) != 1 {
		t.Fatalf("UpdateRecords() = %d; want 1", len(updatedRecords))
	}
	if len(updatedRecords["testZoneID"]) != 1 {
		t.Fatalf("UpdateRecords() = %d; want 1", len(updatedRecords["testZoneID"]))
	}
	if updatedRecords["testZoneID"][0] != "testRecordName" {
		t.Errorf("UpdateRecords() = %s; want testRecordName", updatedRecords["testZoneID"][0])
	}
}

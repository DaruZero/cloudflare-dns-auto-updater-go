package dnsapi

import (
	"bytes"
	"cloudflare-dns-auto-updater-go/cmd/config"
	"cloudflare-dns-auto-updater-go/cmd/mocks"
	"io"
	"net/http"
	"testing"
)

// TestDns_GetZoneID tests the GetZoneId method
func TestDns_GetZoneID(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey: "testAuthKey",
		Email:   "testEmail",
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
	zoneID := dns.GetZoneID("testZoneName")
	// Check the zone id
	if zoneID != "testID" {
		t.Errorf("GetZoneID() = %s; want testID", zoneID)
	}
}

// TestDns_GetCurrentIP tests the GetCurrentIP method
func TestDns_GetCurrentIP(t *testing.T) {
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`"testIP"`))),
			}, nil
		},
	}
	// Create a new dns object
	dns := &CFDNS{
		HTTPClient: mockClient,
	}
	// Get the current IP
	currentIP := dns.GetCurrentIP()
	// Check the current IP
	if currentIP != "testIP" {
		t.Errorf("GetCurrentIP() = %s; want testIP", currentIP)
	}
}

// TestDns_GetRecordsNoRecordID tests the GetRecords method
func TestDns_GetRecordsNoRecordID(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey:  "testAuthKey",
		Email:    "testEmail",
		RecordID: "",
	}
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			json := `{"success":true,"errors":[],"messages":[],"result":[{"id":"testID", "name": "testRecordName", "type": "A", "content": "testContent"}]}`
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
	// Get the records
	records := dns.GetRecords()
	// Check the records
	if len(records) != 1 {
		t.Errorf("GetRecords() = %d; want 1", len(records))
	}
	if records[0].ID != "testID" {
		t.Errorf("GetRecords() = %s; want testID", records[0].ID)
	}
	if records[0].Name != "testRecordName" {
		t.Errorf("GetRecords() = %s; want testRecordName", records[0].Name)
	}
	if records[0].Type != "A" {
		t.Errorf("GetRecords() = %s; want A", records[0].Type)
	}
	if records[0].Content != "testContent" {
		t.Errorf("GetRecords() = %s; want testContent", records[0].Content)
	}
}

// TestDns_GetRecordsWithRecordID tests the GetRecords method
func TestDns_GetRecordsWithRecordID(t *testing.T) {
	// Create a new config object
	cfg := &config.Config{
		AuthKey:  "testAuthKey",
		Email:    "testEmail",
		RecordID: "testID",
	}
	// Create a new mock client
	mockClient := &mocks.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			json := `{"success":true,"errors":[],"messages":[],"result":[{"id":"testID", "name": "testRecordName", "type": "A", "content": "testContent"}]}`
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
	// Get the records
	records := dns.GetRecords()
	// Check the records
	if len(records) != 1 {
		t.Errorf("GetRecords() = %d; want 1", len(records))
	}
	if records[0].ID != "testID" {
		t.Errorf("GetRecords() = %s; want testID", records[0].ID)
	}
	if records[0].Name != "testRecordName" {
		t.Errorf("GetRecords() = %s; want testRecordName", records[0].Name)
	}
	if records[0].Type != "A" {
		t.Errorf("GetRecords() = %s; want A", records[0].Type)
	}
	if records[0].Content != "testContent" {
		t.Errorf("GetRecords() = %s; want testContent", records[0].Content)
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
		Records: []Record{
			{
				ID:      "testID",
				Name:    "testRecordName",
				Type:    "A",
				Content: "testIPToChange",
			},
		},
	}
	// Update the records
	updatedRecords, updated := dns.UpdateRecords()

	// Check the updated records
	if len(updatedRecords) != 1 {
		t.Errorf("UpdateRecords() = %d; want 1", len(updatedRecords))
	}
	if updatedRecords[0] != "testRecordName" {
		t.Errorf("UpdateRecords() = %s; want testRecordName", updatedRecords[0])
	}
	// Check if the records were updated
	if !updated {
		t.Errorf("UpdateRecords() = %t; want true", updated)
	}
}

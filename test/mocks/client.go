package mocks

import (
	"fmt"
	"net/http"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func mockZone(i int) *Zone {
	zoneName := fmt.Sprintf("%d.example.com", i)
	ownerName := "Test Account"

	return &Zone{
		ID:      mockID(zoneName),
		Name:    zoneName,
		DevMode: 0,
		OriginalNS: []string{
			"linda.ns.cloudflare.com",
			"merlin.ns.cloudflare.com",
		},
		OriginalRegistrar: "cloudflare, inc. (id: 1910)",
		OriginalDNSHost:   "",
		CreatedOn:         mustParseTime("2021-07-28T05:06:20.736244Z"),
		ModifiedOn:        mustParseTime("2021-07-28T05:06:20.736244Z"),
		NameServers: []string{
			"abby.ns.cloudflare.com",
			"noel.ns.cloudflare.com",
		},
		Owner: Owner{
			ID:        mockID(ownerName),
			Email:     "",
			Name:      ownerName,
			OwnerType: "organization",
		},
		Permissions: []string{
			"#access:read",
			"#analytics:read",
			"#auditlogs:read",
			"#billing:read",
			"#dns_records:read",
			"#lb:read",
			"#legal:read",
			"#logs:read",
			"#member:read",
			"#organization:read",
			"#ssl:read",
			"#stream:read",
			"#subscription:read",
			"#waf:read",
			"#webhooks:read",
			"#worker:read",
			"#zone:read",
			"#zone_settings:read",
		},
		Plan: ZonePlan{
			ZonePlanCommon: ZonePlanCommon{
				ID:       "0feeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
				Name:     "Free Website",
				Currency: "USD",
			},
			IsSubscribed:      false,
			CanSubscribe:      false,
			LegacyID:          "free",
			LegacyDiscount:    false,
			ExternallyManaged: false,
		},
		PlanPending: ZonePlan{
			ZonePlanCommon: ZonePlanCommon{
				ID: "",
			},
			IsSubscribed:      false,
			CanSubscribe:      false,
			LegacyID:          "",
			LegacyDiscount:    false,
			ExternallyManaged: false,
		},
		Status: "active",
		Paused: false,
		Type:   "full",
		Host: struct {
			Name    string
			Website string
		}{
			Name:    "",
			Website: "",
		},
		VanityNS:    nil,
		Betas:       nil,
		DeactReason: "",
		Meta: ZoneMeta{
			PageRuleQuota:     3,
			WildcardProxiable: false,
			PhishingDetected:  false,
		},
		Account: Account{
			ID:   mockID(ownerName),
			Name: ownerName,
		},
		VerificationKey: "",
	}
}

func mockZonesResponse(total, page, start, count int) *ZonesResponse {
	zones := make([]Zone, count)
	for i := range zones {
		zones[i] = *mockZone(start + i)
	}

	return &ZonesResponse{
		Result: zones,
		ResultInfo: ResultInfo{
			Page:       page,
			PerPage:    50,
			TotalPages: (total + 49) / 50,
			Count:      count,
			Total:      total,
		},
		Response: Response{
			Success:  true,
			Errors:   []ResponseInfo{},
			Messages: []ResponseInfo{},
		},
	}
}

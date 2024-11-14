package broker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	appmodel "open-bos/app/model"
	"strings"
	"testing"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/stretchr/testify/assert"
)

// TestFetchOntology tests the FetchOntology function using a test HTTP server.
func TestFetchOntology(t *testing.T) {
	// Create a test HTTP server to simulate API responses
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Handle token request
		case strings.Contains(r.URL.Path, "/oauth2/v2.0/token"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"access_token": "test-token"}`)
		// Handle getOntologyVersion request
		case strings.Contains(r.URL.Path, "/api/v1/core/application/data/version"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `2`)
		// Handle getOntology request
		case strings.Contains(r.URL.Path, "/api/v1/core/application/data"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{
				"settings": {"version": 2},
				"assetTemplates": [{"id": "asset-template-1", "name": "Temperature Sensor"}],
				"dataTypes": [{"id": "datatype-1", "format": "float", "name": "Temperature", "unitId": "unit-1"}],
				"units": [{"id": "unit-1", "symbol": "°C"}],
				"datapointTemplates": [{"id": "datapoint-template-1", "name": "Temperature", "assetTemplateId": "asset-template-1", "typeId": "datatype-1", "direction": "feedback"}],
				"assets": [{"id": "asset-1", "name": "Sensor 1", "templateId": "asset-template-1"}],
				"datapoints": [{"id": "datapoint-1", "templateId": "datapoint-template-1", "assetId": "asset-1"}]
			}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Prepare configuration
	config := appmodel.Configuration{
		Id:              1,
		Gwid:            "test-gwid",
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		AppPublicAPIURL: ts.URL,
		OntologyVersion: 1, // Previous version
	}

	// Create a custom newOpenBOSClient function for testing
	originalNewOpenBOSClient := newOpenBOSClient
	newOpenBOSClient = func(gatewayID, clientID, clientSecret, webhookURL, baseURL, tokenURL string) (*openBOSClient, error) {
		client := &openBOSClient{
			gatewayID:    gatewayID,
			httpClient:   ts.Client(),
			clientID:     clientID,
			clientSecret: clientSecret,
			webhookURL:   webhookURL,
			baseURL:      ts.URL,
			tokenURL:     ts.URL + "/oauth2/v2.0/token",
		}

		// Bypass authentication for testing
		client.accessToken = "test-token"

		return client, nil
	}
	defer func() { newOpenBOSClient = originalNewOpenBOSClient }()

	// Call FetchOntology
	ontologyVersion, assetTypes, rootAsset, err := FetchOntology(config)
	if err != nil {
		t.Fatalf("FetchOntology returned error: %v", err)
	}

	// Assertions
	assert.Equal(t, int32(2), ontologyVersion, "Ontology version should be updated")

	expectedAssetType := api.AssetType{
		Name: "open_bos_asset-template-1",
		Translation: *api.NewNullableTranslation(&api.Translation{
			En: common.Ptr("OpenBOS Temperature Sensor"),
		}),
		Attributes: []api.AssetTypeAttribute{
			{
				Name:    "Temperature",
				Subtype: api.SUBTYPE_INPUT,
				Unit:    *api.NewNullableString(common.Ptr("°C")),
			},
			{
				Name:      masterPropertyAttribute,
				Subtype:   api.SUBTYPE_PROPERTY,
				IsDigital: *api.NewNullableBool(common.Ptr(true)),
				Map: []map[string]interface{}{
					{
						"value": -1,
						"map":   "Not available",
					},
					{
						"value": 0,
						"map":   "Slave",
					},
					{
						"value": 1,
						"map":   "Master",
					},
				},
			},
		},
	}

	if len(assetTypes) != 2 {
		t.Fatalf("Expected 2 asset types, got %d", len(assetTypes))
	}

	// The second asset type is the root asset type.
	// For simplicity, we'll check the first one.
	assert.Equal(t, expectedAssetType, assetTypes[0])

	// Check that root asset is correctly built
	if rootAsset.Name != "OpenBOS" {
		t.Errorf("Expected root asset name 'OpenBOS', got '%s'", rootAsset.Name)
	}

	// Check that the asset hierarchy is correctly built
	if len(rootAsset.FunctionalChildrenSlice) != 1 {
		t.Errorf("Expected 1 child asset, got %d", len(rootAsset.FunctionalChildrenSlice))
	}

	childAsset := rootAsset.FunctionalChildrenSlice[0]
	if childAsset.Name != "Sensor 1" {
		t.Errorf("Expected child asset name 'Sensor 1', got '%s'", childAsset.Name)
	}
}

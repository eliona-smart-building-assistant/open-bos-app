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

// TestFetchOntologyWithComplexDataTypes tests handling of complex data types.
func TestFetchOntologyWithComplexDataTypes(t *testing.T) {
	// Prepare configuration
	config := appmodel.Configuration{
		Id:              1,
		Gwid:            "test-gwid",
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		AppPublicAPIURL: "http://test-api-url",
		OntologyVersion: 1, // Previous version
	}

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
				"dataTypes": [
					{
						"id": "11111111-1111-1111-1111-111111111111",
						"tag": "bos:point:temp:setpoint_cool",
						"format": "double",
						"unitId": "Temperature_DegreesCelsius",
						"min": -999999.0,
						"max": 99999.99
					},
					{
						"id": "22222222-2222-2222-2222-222222222222",
						"name": "HVAC_Status",
						"tag": "bos:standardtype:hvacstatus",
						"format": "enumeration",
						"enums": {
							"0": "Auto",
							"1": "Comfort",
							"2": "Standby",
							"3": "Economy",
							"4": "Building Protection"
						}
					},
					{
						"id": "00000000-0000-0000-0000-000000000000",
						"tag": "bos:point:something",
						"format": "complex",
						"fields": [
							{
								"name": "myFirstField",
								"typeId": "11111111-1111-1111-1111-111111111111"
							},
							{
								"name": "mySecondField",
								"typeId": "22222222-2222-2222-2222-222222222222"
							}
						]
					},
					{
						"id": "33333333-3333-3333-3333-333333333333",
						"tag": "bos:point:verycomplex",
						"format": "complex",
						"fields": [
							{
								"name": "complexInComplex",
								"typeId": "00000000-0000-0000-0000-000000000000"
							},
							{
								"name": "mySecondField",
								"typeId": "22222222-2222-2222-2222-222222222222"
							}
						]
					}
				],
				"units": [
					{
						"id": "Temperature_DegreesCelsius",
						"symbol": "°C"
					}
				],
				"assetTemplates": [
					{
						"id": "asset-template-1",
						"name": "Complex Asset"
					}
				],
				"datapointTemplates": [
					{
						"id": "datapoint-template-complex",
						"name": "Complex DataPoint",
						"assetTemplateId": "asset-template-1",
						"typeId": "33333333-3333-3333-3333-333333333333",
						"direction": "feedback"
					},
					{
						"id": "datapoint-template-simple-with-name",
						"name": "Complex DataPoint",
						"assetTemplateId": "asset-template-1",
						"typeId": "22222222-2222-2222-2222-222222222222",
						"direction": "feedback"
					}
				],
				"assets": [
					{
						"id": "asset-1",
						"name": "Complex Asset Instance",
						"templateId": "asset-template-1"
					}
				],
				"datapoints": [
					{
						"id": "datapoint-1",
						"templateId": "datapoint-template-complex",
						"assetId": "asset-1"
					},
					{
						"id": "datapoint-1",
						"templateId": "datapoint-template-simple-with-name",
						"assetId": "asset-1"
					}
				]
			}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Override newOpenBOSClient in the test
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
			accessToken:  "test-token",
		}
		return client, nil
	}
	defer func() { newOpenBOSClient = originalNewOpenBOSClient }()

	// Call the function under test
	ontologyVersion, assetTypes, _, err := FetchOntology(config)
	if err != nil {
		t.Fatalf("FetchOntology returned error: %v", err)
	}

	// Assertions
	assert.Equal(t, int32(2), ontologyVersion, "Ontology version should be updated")

	// Find the asset type corresponding to "asset-template-1"
	var assetType api.AssetType
	for _, at := range assetTypes {
		if at.Name == "open_bos_asset-template-1" {
			assetType = at
			break
		}
	}

	// Expected attributes after unwrapping complex data types
	expectedAttributes := []api.AssetTypeAttribute{
		{
			Name:    "HVAC_Status",
			Subtype: api.SUBTYPE_INPUT,
			Map: []map[string]interface{}{
				{"value": "0", "map": "Auto"},
				{"value": "1", "map": "Comfort"},
				{"value": "2", "map": "Standby"},
				{"value": "3", "map": "Economy"},
				{"value": "4", "map": "Building Protection"},
			},
		},
		{
			Name:    "33333333-3333-3333-3333-333333333333.complexInComplex.myFirstField",
			Subtype: api.SUBTYPE_INPUT,
			Unit:    *api.NewNullableString(common.Ptr("°C")),
			Min:     *api.NewNullableFloat64(common.Ptr(-999999.0)),
			Max:     *api.NewNullableFloat64(common.Ptr(99999.99)),
		},
		{
			Name:    "33333333-3333-3333-3333-333333333333.complexInComplex.mySecondField",
			Subtype: api.SUBTYPE_INPUT,
			Map: []map[string]interface{}{
				{"value": "0", "map": "Auto"},
				{"value": "1", "map": "Comfort"},
				{"value": "2", "map": "Standby"},
				{"value": "3", "map": "Economy"},
				{"value": "4", "map": "Building Protection"},
			},
		},
		{
			Name:    "33333333-3333-3333-3333-333333333333.mySecondField",
			Subtype: api.SUBTYPE_INPUT,
			Map: []map[string]interface{}{
				{"value": "0", "map": "Auto"},
				{"value": "1", "map": "Comfort"},
				{"value": "2", "map": "Standby"},
				{"value": "3", "map": "Economy"},
				{"value": "4", "map": "Building Protection"},
			},
		},
		{
			Name:      masterPropertyAttribute,
			Subtype:   api.SUBTYPE_PROPERTY,
			IsDigital: *api.NewNullableBool(common.Ptr(true)),
			Map: []map[string]interface{}{
				{"value": -1, "map": "Not available"},
				{"value": 0, "map": "Slave"},
				{"value": 1, "map": "Master"},
			},
		},
	}

	// Check that the attributes match expected
	assert.Equal(t, len(expectedAttributes), len(assetType.Attributes), "Attribute count mismatch")

	for _, expectedAttr := range expectedAttributes {
		found := false
		for _, actualAttr := range assetType.Attributes {
			if expectedAttr.Name == actualAttr.Name {
				// Compare other fields
				assert.Equal(t, expectedAttr.Name, actualAttr.Name, "Name mismatch for %s", expectedAttr.Name)
				assert.Equal(t, expectedAttr.Subtype, actualAttr.Subtype, "Subtype mismatch for %s", expectedAttr.Name)
				assert.Equal(t, expectedAttr.Unit, actualAttr.Unit, "Unit mismatch for %s", expectedAttr.Name)
				assert.Equal(t, expectedAttr.Min, actualAttr.Min, "Min mismatch for %s", expectedAttr.Name)
				assert.Equal(t, expectedAttr.Max, actualAttr.Max, "Max mismatch for %s", expectedAttr.Name)
				assert.Equal(t, expectedAttr.IsDigital, actualAttr.IsDigital, "IsDigital mismatch for %s", expectedAttr.Name)

				// Use assert.ElementsMatch to compare Map slices regardless of order
				assert.ElementsMatch(t, expectedAttr.Map, actualAttr.Map, "Map mismatch for %s", expectedAttr.Name)

				found = true
				break
			}
		}
		assert.True(t, found, "Expected attribute %s not found", expectedAttr.Name)
	}
}

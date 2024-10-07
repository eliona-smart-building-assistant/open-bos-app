package broker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const baseURL = "https://api.buildings.ability.abb/buildings/openbos/apiproxy/v1"
const mockURL = "http://localhost:5000"

type openBOSClient struct {
	gatewayID    string
	httpClient   *http.Client
	accessToken  string
	clientID     string
	clientSecret string
}

func newOpenBOSClient(gatewayID, clientID, clientSecret string) (*openBOSClient, error) {
	client := &openBOSClient{
		gatewayID:    gatewayID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	if err := client.authenticateWithClientCredentials(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *openBOSClient) authenticateWithClientCredentials() error {
	tokenURL := "https://login.microsoftonline.com/372ee9e0-9ce0-4033-a64a-c07073a91ecd/oauth2/v2.0/token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", "api://openbos/.default")

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to obtain access token, status code: %d body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return fmt.Errorf("unmarshalling: %v", err)
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return errors.New("invalid token response")
	}

	c.accessToken = accessToken
	return nil
}

func (c *openBOSClient) doRequest(method, endpoint string, queryParams url.Values, body interface{}, result interface{}) error {
	url := endpoint
	if queryParams != nil && len(queryParams) > 0 {
		url += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %v", err)
		}
	}

	return nil
}

func (c *openBOSClient) doMockRequest(method, endpoint string, queryParams url.Values, body interface{}, result interface{}) error {
	url := endpoint
	if queryParams != nil && len(queryParams) > 0 {
		url += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %v", err)
		}
	}

	return nil
}

type ontologyDTO struct {
	Settings           ontologySettingsDTO               `json:"settings"`
	AssetTemplates     []ontologyAssetOrSpaceTemplateDTO `json:"assetTemplates,omitempty"`
	SpaceTemplates     []ontologyAssetOrSpaceTemplateDTO `json:"spaceTemplates,omitempty"`
	Units              []ontologyUnitDTO                 `json:"units,omitempty"`
	DataTypes          []ontologyDataTypeDTO             `json:"dataTypes,omitempty"`
	DatapointTemplates []ontologyDatapointTemplateDTO    `json:"datapointTemplates,omitempty"`
	PropertyTemplates  []ontologyPropertyTemplateDTO     `json:"propertyTemplates,omitempty"`
	Assets             []ontologyAssetDTO                `json:"assets,omitempty"`
	Spaces             []ontologySpaceDTO                `json:"spaces,omitempty"`
	Datapoints         []ontologyDatapointDTO            `json:"datapoints,omitempty"`
	Properties         []ontologyPropertyDTO             `json:"properties,omitempty"`
	Orgs               []ontologyOrganisationDTO         `json:"orgs,omitempty"`
	Users              []ontologyUserDTO                 `json:"users,omitempty"`
}

type ontologySettingsDTO struct {
	Version int64 `json:"version"`
}

type ontologyAssetOrSpaceTemplateDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Tags          []string `json:"tags,omitempty"`
	Icon          string   `json:"icon,omitempty"`
	IconFillColor string   `json:"iconFillColor,omitempty"`
}

type ontologyUnitDTO struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
}

type ontologyDataTypeDTO struct {
	Format string                     `json:"format"`
	ID     string                     `json:"id"`
	Name   string                     `json:"name"`
	Tags   []string                   `json:"tags,omitempty"`
	UnitID string                     `json:"unitId,omitempty"`
	Fields []ontologyDataTypeFieldDTO `json:"fields,omitempty"`
	Min    *float64                   `json:"min,omitempty"`
	Max    *float64                   `json:"max,omitempty"`
	Enums  map[string]string          `json:"enums,omitempty"`
}

type ontologyDataTypeFieldDTO struct {
	Name   string `json:"name"`
	TypeID string `json:"typeId"`
}

type ontologyDatapointTemplateDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	AssetTemplateID string   `json:"assetTemplateId,omitempty"`
	SpaceTemplateID string   `json:"spaceTemplateId,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	TypeID          string   `json:"typeId,omitempty"`
	Direction       string   `json:"direction"`
	Perpetual       bool     `json:"perpetual"`
}

type ontologyPropertyTemplateDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	SpaceTemplateID string   `json:"spaceTemplateId,omitempty"`
	AssetTemplateID string   `json:"assetTemplateId,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	TypeID          string   `json:"typeId,omitempty"`
}

type ontologyAssetDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	TemplateID string   `json:"templateId"`
	Tags       []string `json:"tags,omitempty"`
}

type ontologySpaceDTO struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	ParentID   string                  `json:"parentId,omitempty"`
	TemplateID string                  `json:"templateId"`
	Tags       []string                `json:"tags,omitempty"`
	Assets     []ontologySpaceAssetDTO `json:"assets,omitempty"`

	children []ontologySpaceDTO
}

type ontologySpaceAssetDTO struct {
	ID     string `json:"id"`
	Master bool   `json:"master"`
}

type ontologyDatapointDTO struct {
	ID         string `json:"id"`
	TemplateID string `json:"templateId"`
	AssetID    string `json:"assetId,omitempty"`
	SpaceID    string `json:"spaceId,omitempty"`
}

type ontologyPropertyDTO struct {
	ID         string      `json:"id"`
	TemplateID string      `json:"templateId"`
	SpaceID    string      `json:"spaceId,omitempty"`
	AssetID    string      `json:"assetId,omitempty"`
	Value      interface{} `json:"value,omitempty"`
}

type ontologyOrganisationDTO struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

type ontologyUserDTO struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Initials string   `json:"initials,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Email    string   `json:"email,omitempty"`
	OrgID    string   `json:"orgId,omitempty"`
}

// getOntology retrieves the complete ontology of the edge.
func (c *openBOSClient) getOntology() (*ontologyDTO, error) {
	//endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full", baseURL, c.gatewayID)
	endpoint := fmt.Sprintf("%s/api/v1/ontology/full", mockURL)

	var ontology ontologyDTO
	if err := c.doMockRequest("GET", endpoint, nil, nil, &ontology); err != nil {
		return nil, err
	}

	return &ontology, nil
}

// getOntologyVersion retrieves the current version of the edge data.
func (c *openBOSClient) getOntologyVersion() (int64, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/version", baseURL, c.gatewayID)

	var version int64
	if err := c.doRequest("GET", endpoint, nil, nil, &version); err != nil {
		return 0, err
	}

	return version, nil
}

type propertyDTO struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Icon             string        `json:"icon,omitempty"`
	IconFillColor    string        `json:"iconFillColor,omitempty"`
	TemplateID       string        `json:"templateId,omitempty"`
	Tags             []string      `json:"tags,omitempty"`
	PropertyZoneType string        `json:"propertyZoneType,omitempty"`
	IsExternal       bool          `json:"isExternal"`
	Authorized       bool          `json:"authorized"`
	ParentIDs        []string      `json:"parentIds,omitempty"`
	ChildrenIDs      []string      `json:"childrenIds,omitempty"`
	AllChildrenCount int32         `json:"allChildrenCount"`
	HasMapView       bool          `json:"hasMapView"`
	DisplayIndex     int32         `json:"displayIndex"`
	Datapoints       []interface{} `json:"datapoints,omitempty"`
	Properties       []interface{} `json:"properties,omitempty"`
	Address1         string        `json:"address1,omitempty"`
	Address2         string        `json:"address2,omitempty"`
	Town             string        `json:"town,omitempty"`
	Country          string        `json:"country,omitempty"`
	State            string        `json:"state,omitempty"`
	SurfaceGross     string        `json:"surfaceGross,omitempty"`
	SurfaceNet       string        `json:"surfaceNet,omitempty"`
	People           string        `json:"people,omitempty"`
	Image            string        `json:"image,omitempty"`
	GPSLat           string        `json:"gpsLat,omitempty"`
	GPSLon           string        `json:"gpsLon,omitempty"`
	Buildings        string        `json:"buildings,omitempty"`
	Tenants          string        `json:"tenants,omitempty"`
	Owners           string        `json:"owners,omitempty"`
	PropertyManagers string        `json:"propertyManagers,omitempty"`
	FacilityManagers string        `json:"facilityManagers,omitempty"`
	Visitors         string        `json:"visitors,omitempty"`
	Administrators   string        `json:"administrators,omitempty"`
	Devices          string        `json:"devices,omitempty"`
}

// getProperty retrieves the Property (Site) description.
func (c *openBOSClient) getProperty() (*propertyDTO, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/property", baseURL, c.gatewayID)

	var property propertyDTO
	if err := c.doRequest("GET", endpoint, nil, nil, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

type ontologyFullLiveAlarmDTO struct {
	DataPointInstanceID string      `json:"dataPointInstanceId"`
	SessionID           string      `json:"sessionId"`
	Name                string      `json:"name"`
	Description         string      `json:"description"`
	Trigger             string      `json:"trigger"`
	Active              bool        `json:"active"`
	Acked               bool        `json:"acked"`
	Closed              bool        `json:"closed"`
	TimeStamp           string      `json:"timeStamp"`
	Quality             string      `json:"quality"`
	Value               interface{} `json:"value"`
	AckedBy             string      `json:"ackedBy"`
	Comment             string      `json:"comment"`
	NeedAcknowledge     bool        `json:"needAcknowledge"`
	Severity            string      `json:"severity"`
	AssetID             string      `json:"assetId"`
	SpaceID             string      `json:"spaceId"`
	AssetName           string      `json:"assetName"`
	SpaceName           string      `json:"spaceName"`
	DatapointName       string      `json:"datapointName"`
	UnitSymbol          string      `json:"unitSymbol"`
	Tags                []string    `json:"tags,omitempty"`
}

func (c *openBOSClient) getLiveAlarms(timestamp string) ([]ontologyFullLiveAlarmDTO, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/livealarm", baseURL, c.gatewayID)

	params := url.Values{}
	if timestamp != "" {
		params.Add("timestamp", timestamp)
	}

	var alarms []ontologyFullLiveAlarmDTO
	if err := c.doRequest("GET", endpoint, params, nil, &alarms); err != nil {
		return nil, err
	}

	return alarms, nil
}

type ontologyAlarmAckDTO struct {
	SessionID string `json:"sessionId,omitempty"`
	AckedBy   string `json:"ackedBy,omitempty"`
	AckedByID string `json:"ackedById,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func (c *openBOSClient) ackAlarm(ack ontologyAlarmAckDTO) error {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/livealarm/ack", baseURL, c.gatewayID)

	if err := c.doRequest("POST", endpoint, nil, ack, nil); err != nil {
		return err
	}

	return nil
}

type dataPoint struct {
	ID            string
	Name          string
	Tags          []string
	Direction     string
	TypeID        string
	DisplayUnitID *string
	Min           *float64
	Max           *float64
	Enums         map[string]string
}

type property struct {
	ID            string
	Name          string
	Tags          []string
	TypeID        string
	DisplayUnitID *string
	Min           *float64
	Max           *float64
}

type assetTemplate struct {
	ID         string
	Name       string
	Tags       []string
	Properties []property
	Datapoints []dataPoint
}

func (ontology ontologyDTO) getAssetTemplates() []assetTemplate {
	datapointTemplateMap := make(map[string][]ontologyDatapointTemplateDTO)
	for _, dt := range ontology.DatapointTemplates {
		if dt.AssetTemplateID == "" {
			log.Warn("client", "datapoint template %s has no template ID", dt.ID)
			continue
		}
		datapointTemplateMap[dt.AssetTemplateID] = append(datapointTemplateMap[dt.AssetTemplateID], dt)
	}

	propertyTemplateMap := make(map[string][]ontologyPropertyTemplateDTO)
	for _, pt := range ontology.PropertyTemplates {
		switch {
		case pt.AssetTemplateID != "" && pt.SpaceTemplateID != "":
			log.Warn("client", "property template %s has both asset and space template ID", pt.ID)
			continue
		case pt.AssetTemplateID != "":
			propertyTemplateMap[pt.AssetTemplateID] = append(propertyTemplateMap[pt.AssetTemplateID], pt)
		case pt.SpaceTemplateID != "":
			propertyTemplateMap[pt.SpaceTemplateID] = append(propertyTemplateMap[pt.SpaceTemplateID], pt)
		default:
			log.Warn("client", "property template %s has neither asset nor space template ID", pt.ID)
			continue
		}
	}

	dataTypeMap := make(map[string]ontologyDataTypeDTO)
	for _, dt := range ontology.DataTypes {
		dataTypeMap[dt.ID] = dt
	}

	unitMap := make(map[string]string)
	for _, unit := range ontology.Units {
		unitMap[unit.ID] = unit.Symbol
	}

	var assetTemplates []assetTemplate
	for _, at := range append(ontology.AssetTemplates, ontology.SpaceTemplates...) {
		assetTemplate := assetTemplate{
			ID:         at.ID,
			Name:       at.Name,
			Tags:       at.Tags,
			Datapoints: []dataPoint{},
			Properties: []property{},
		}

		for _, dt := range datapointTemplateMap[at.ID] {
			dataType := getDataType(dt.TypeID, dataTypeMap)
			var min, max *float64
			var displayUnit *string
			var enums map[string]string
			if dataType != nil {
				min = dataType.Min
				max = dataType.Max
				enums = dataType.Enums
				displayUnit = getDisplayUnitID(*dataType, unitMap)
			}
			dataPoint := dataPoint{
				ID:            dt.ID,
				Name:          dt.Name,
				Tags:          dt.Tags,
				TypeID:        dt.TypeID,
				Min:           min,
				Max:           max,
				Enums:         enums,
				Direction:     dt.Direction,
				DisplayUnitID: displayUnit,
			}
			assetTemplate.Datapoints = append(assetTemplate.Datapoints, dataPoint)
		}

		for _, pt := range propertyTemplateMap[at.ID] {
			dataType := getDataType(pt.TypeID, dataTypeMap)
			var min, max *float64
			var displayUnit *string
			if dataType != nil {
				min = dataType.Min
				max = dataType.Max
				displayUnit = getDisplayUnitID(*dataType, unitMap)
			}
			property := property{
				ID:            pt.ID,
				Name:          pt.Name,
				Tags:          pt.Tags,
				TypeID:        pt.TypeID,
				Min:           min,
				Max:           max,
				DisplayUnitID: displayUnit,
			}
			assetTemplate.Properties = append(assetTemplate.Properties, property)
		}

		assetTemplates = append(assetTemplates, assetTemplate)
	}

	return assetTemplates
}

func getDataType(typeID string, dataTypeMap map[string]ontologyDataTypeDTO) *ontologyDataTypeDTO {
	dataType, exists := dataTypeMap[typeID]
	if !exists {
		log.Warn("client", "type %s not found", typeID)
		return nil
	}
	return &dataType
}

func getDisplayUnitID(dataType ontologyDataTypeDTO, unitMap map[string]string) *string {
	unitSymbol, ok := unitMap[dataType.UnitID]
	if !ok {
		log.Warn("client", "unit %s not found", dataType.UnitID)
		return nil
	}
	return &unitSymbol
}

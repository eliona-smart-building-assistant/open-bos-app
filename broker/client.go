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
)

const baseURL = "https://api.buildings.ability.abb/buildings/openbos/apiproxy/v1"

type OpenBOSClient struct {
	GatewayID    string
	HTTPClient   *http.Client
	AccessToken  string
	clientID     string
	clientSecret string
}

func NewOpenBOSClient(gatewayID, clientID, clientSecret string) (*OpenBOSClient, error) {
	client := &OpenBOSClient{
		GatewayID:    gatewayID,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	if err := client.authenticateWithClientCredentials(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *OpenBOSClient) authenticateWithClientCredentials() error {
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

	resp, err := c.HTTPClient.Do(req)
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

	c.AccessToken = accessToken
	return nil
}

func (c *OpenBOSClient) doRequest(method, endpoint string, queryParams url.Values, body interface{}, result interface{}) error {
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

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
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

type OntologyDTO struct {
	Settings           OntologySettingsDTO            `json:"settings"`
	AssetTemplates     []OntologyAssetTemplateDTO     `json:"assetTemplates,omitempty"`
	SpaceTemplates     []OntologySpaceTemplateDTO     `json:"spaceTemplates,omitempty"`
	Units              []OntologyUnitDTO              `json:"units,omitempty"`
	DataTypes          []OntologyDataTypeDTO          `json:"dataTypes,omitempty"`
	DatapointTemplates []OntologyDatapointTemplateDTO `json:"datapointTemplates,omitempty"`
	PropertyTemplates  []OntologyPropertyTemplateDTO  `json:"propertyTemplates,omitempty"`
	Assets             []OntologyAssetDTO             `json:"assets,omitempty"`
	Spaces             []OntologySpaceDTO             `json:"spaces,omitempty"`
	Datapoints         []OntologyDatapointDTO         `json:"datapoints,omitempty"`
	Properties         []OntologyPropertyDTO          `json:"properties,omitempty"`
	Orgs               []OntologyOrganisationDTO      `json:"orgs,omitempty"`
	Users              []OntologyUserDTO              `json:"users,omitempty"`
}

type OntologySettingsDTO struct {
	Version int64 `json:"version"`
}

type OntologyAssetTemplateDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Tags          []string `json:"tags,omitempty"`
	Icon          string   `json:"icon,omitempty"`
	IconFillColor string   `json:"iconFillColor,omitempty"`
}

type OntologySpaceTemplateDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Tags          []string `json:"tags,omitempty"`
	Icon          string   `json:"icon,omitempty"`
	IconFillColor string   `json:"iconFillColor,omitempty"`
}

type OntologyUnitDTO struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
}

type OntologyDataTypeDTO struct {
	Format string                     `json:"format"`
	ID     string                     `json:"id"`
	Name   string                     `json:"name"`
	Tags   []string                   `json:"tags,omitempty"`
	UnitID string                     `json:"unitId,omitempty"`
	Fields []OntologyDataTypeFieldDTO `json:"fields,omitempty"`
	Min    *float64                   `json:"min,omitempty"`
	Max    *float64                   `json:"max,omitempty"`
	Enums  map[string]string          `json:"enums,omitempty"`
}

type OntologyDataTypeFieldDTO struct {
	Name   string `json:"name"`
	TypeID string `json:"typeId"`
}

type OntologyDatapointTemplateDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	AssetTemplateID string   `json:"assetTemplateId,omitempty"`
	SpaceTemplateID string   `json:"spaceTemplateId,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	TypeID          string   `json:"typeId,omitempty"`
	Perpetual       bool     `json:"perpetual"`
}

type OntologyPropertyTemplateDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	SpaceTemplateID string   `json:"spaceTemplateId,omitempty"`
	AssetTemplateID string   `json:"assetTemplateId,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	TypeID          string   `json:"typeId,omitempty"`
}

type OntologyAssetDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	TemplateID string   `json:"templateId"`
	Tags       []string `json:"tags,omitempty"`
}

type OntologySpaceDTO struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	ParentID   string                  `json:"parentId,omitempty"`
	TemplateID string                  `json:"templateId"`
	Tags       []string                `json:"tags,omitempty"`
	Assets     []OntologySpaceAssetDTO `json:"assets,omitempty"`
}

type OntologySpaceAssetDTO struct {
	ID     string `json:"id"`
	Master bool   `json:"master"`
}

type OntologyDatapointDTO struct {
	ID         string `json:"id"`
	TemplateID string `json:"templateId"`
	AssetID    string `json:"assetId,omitempty"`
	SpaceID    string `json:"spaceId,omitempty"`
}

type OntologyPropertyDTO struct {
	ID         string      `json:"id"`
	TemplateID string      `json:"templateId"`
	SpaceID    string      `json:"spaceId,omitempty"`
	AssetID    string      `json:"assetId,omitempty"`
	Value      interface{} `json:"value,omitempty"`
}

type OntologyOrganisationDTO struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

type OntologyUserDTO struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Initials string   `json:"initials,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Email    string   `json:"email,omitempty"`
	OrgID    string   `json:"orgId,omitempty"`
}

// GetOntology retrieves the complete ontology of the edge.
func (c *OpenBOSClient) GetOntology(types []string) (*OntologyDTO, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full", baseURL, c.GatewayID)

	params := url.Values{}
	for _, t := range types {
		params.Add("types", t)
	}

	var ontology OntologyDTO
	err := c.doRequest("GET", endpoint, params, nil, &ontology)
	if err != nil {
		return nil, err
	}

	return &ontology, nil
}

// GetOntologyVersion retrieves the current version of the edge data.
func (c *OpenBOSClient) GetOntologyVersion() (int64, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/version", baseURL, c.GatewayID)

	var version int64
	err := c.doRequest("GET", endpoint, nil, nil, &version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

type PropertyDTO struct {
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

// GetProperty retrieves the Property (Site) description.
func (c *OpenBOSClient) GetProperty() (*PropertyDTO, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/property", baseURL, c.GatewayID)

	var property PropertyDTO
	err := c.doRequest("GET", endpoint, nil, nil, &property)
	if err != nil {
		return nil, err
	}

	return &property, nil
}

type OntologyFullLiveAlarmDTO struct {
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

func (c *OpenBOSClient) GetLiveAlarms(timestamp string) ([]OntologyFullLiveAlarmDTO, error) {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/livealarm", baseURL, c.GatewayID)

	params := url.Values{}
	if timestamp != "" {
		params.Add("timestamp", timestamp)
	}

	var alarms []OntologyFullLiveAlarmDTO
	err := c.doRequest("GET", endpoint, params, nil, &alarms)
	if err != nil {
		return nil, err
	}

	return alarms, nil
}

type OntologyAlarmAckDTO struct {
	SessionID string `json:"sessionId,omitempty"`
	AckedBy   string `json:"ackedBy,omitempty"`
	AckedByID string `json:"ackedById,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func (c *OpenBOSClient) AckAlarm(ack OntologyAlarmAckDTO) error {
	endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/livealarm/ack", baseURL, c.GatewayID)

	err := c.doRequest("POST", endpoint, nil, ack, nil)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Deprecate or not?

type DataPoint struct {
	BusUnitId     *string  `json:"busUnitId"`
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Direction     string   `json:"direction"`
	TypeId        string   `json:"typeId"`
	DisplayUnitId *string  `json:"displayUnitId"`
	PublicId      string   `json:"publicId"`
}

type Property struct {
	DefaultValue  any      `json:"defaultValue"`
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Direction     string   `json:"direction"`
	TypeId        string   `json:"typeId"`
	DisplayUnitId *string  `json:"displayUnitId"`
	PublicId      string   `json:"publicId"`
}

type AssetTemplate struct {
	Datapoints         []DataPoint   `json:"datapoints"`
	Properties         []Property    `json:"properties"`
	Usages             []interface{} `json:"usages"`
	Id                 string        `json:"id"`
	Icon               string        `json:"icon"`
	IconFillColor      *string       `json:"iconFillColor"`
	Name               string        `json:"name"`
	Tags               []string      `json:"tags"`
	ParentId           *string       `json:"parentId"`
	Version            string        `json:"version"`
	InstancesCount     int           `json:"instancesCount"`
	PublicId           string        `json:"publicId"`
	IsExternal         bool          `json:"isExternal"`
	SupportMasterSlave bool          `json:"supportMasterSlave"`
	IsDefault          bool          `json:"isDefault"`
}

func (c *OpenBOSClient) getAssetTemplates() ([]AssetTemplate, error) {
	url := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/functionalblocktemplate/details", baseURL, c.GatewayID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(body))
	}

	var templates []AssetTemplate
	if err := json.Unmarshal(body, &templates); err != nil {
		return nil, fmt.Errorf("unmarshalling: %v", err)
	}
	return templates, nil
}

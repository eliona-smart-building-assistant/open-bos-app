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

	"github.com/eliona-smart-building-assistant/go-utils/common"
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
	webhookURL   string
}

func newOpenBOSClient(gatewayID, clientID, clientSecret, webhookURL string) (*openBOSClient, error) {
	client := &openBOSClient{
		gatewayID:    gatewayID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		webhookURL:   webhookURL,
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// TODO: This part may be removed later
	// {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("dumping request: %v", err)
	}
	log.Debug("client", "HTTP Request:\n%s\n", string(dump))

	// DumpRequest may consume the body, so we need to reset it afterward
	if bodyReader != nil {
		if seeker, ok := bodyReader.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		req.Body = io.NopCloser(bodyReader)
	}
	// }

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
	Version int32 `json:"version"`
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

// unwrapComplexType recursively flattens complex datatypes into a slice of simple ones.
func (dt ontologyDataTypeDTO) unwrapComplexType(datatypeComplexMap map[string]ontologyDataTypeDTO) []ontologyDataTypeDTO {
	if len(dt.Fields) == 0 {
		return []ontologyDataTypeDTO{dt}
	}
	var result []ontologyDataTypeDTO
	for _, childReference := range dt.Fields {
		child := datatypeComplexMap[childReference.TypeID]
		if childReference.Name != "" {
			// TODO: Is this the right priority?
			child.Name = childReference.Name
		}
		result = append(result, child.unwrapComplexType(datatypeComplexMap)...)
	}

	return result
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
func (c *openBOSClient) getOntologyVersion() (int32, error) {
	//endpoint := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/full/version", baseURL, c.gatewayID)
	endpoint := fmt.Sprintf("%s/api/v1/ontology/full/version", mockURL)

	var version int32
	if err := c.doRequest("GET", endpoint, nil, nil, &version); err != nil {
		return 0, err
	}

	return version, nil
}

type ontologySubscriptionCreateDTO struct {
	MinSendTime       int32   `json:"minSendTime"`              // Minimum time between two events. To avoid events flushing. Highly recommended. If zero, events will be sent on the fly (NOT recommended). Min value: 1 mn
	MaxSendTime       int32   `json:"maxSendTime"`              // Maximum time between two events. Can be used to ensure the client application that the connection is alive. If nothing must be sent, the edge will send an empty event. Min value: 1 mn
	Timestamp         *string `json:"timestamp,omitempty"`      // UTC date. To receive only datapoints or properties that change since the timestamp.
	WebHookURL        *string `json:"webHookURL,omitempty"`     // URL of the webhook.
	WebHookRetries    int32   `json:"webHookRetries"`           // Interval of retries (in seconds) when an error occurs while sending an event.
	WebHookRetryDelay int32   `json:"webHookRetryDelay"`        // Number of retries when an error occurs while sending an event.
	WebHookLeaseTime  int32   `json:"webHookLeaseTime"`         // Life span of the webhook if the webhook connection is down. If not present, the webhook will never be destroyed.
	WebhookPersist    *bool   `json:"webhookPersist,omitempty"` // If true, the subscription will be kept alive when the edge restarts in the middle of the subscription. If false, the subscription is lost when the edge restarts.
	ContentType       *string `json:"contentType,omitempty"`    // application/json for json (the default) or octet for base64.
}

type subscriptionResultDTO struct {
	ID         *string `json:"id,omitempty"`
	WebHookURL *string `json:"webHookURL,omitempty"`
}

func (c *openBOSClient) subscribeToOntologyChanges(configID int64) (*subscriptionResultDTO, error) {
	endpoint := fmt.Sprintf("%s/api/v1/ontology/full/version/subscribe", mockURL)

	webhookURL, err := url.JoinPath(c.webhookURL, fmt.Sprint(configID), "ontology-version")
	if err != nil {
		return nil, fmt.Errorf("joining URL for subscription: %v", err)
	}

	second := int32(1000)
	minute := 60 * second
	sub := ontologySubscriptionCreateDTO{
		MinSendTime:       5 * minute,
		WebHookURL:        common.Ptr(webhookURL),
		WebHookRetries:    3,
		WebHookRetryDelay: 5 * second,
		WebHookLeaseTime:  60 * minute,
		WebhookPersist:    common.Ptr(true),
		ContentType:       common.Ptr("application/json"),
	}

	var result subscriptionResultDTO
	if err := c.doMockRequest("POST", endpoint, nil, sub, &result); err != nil {
		return nil, fmt.Errorf("failed to subscribe to ontology changes: %v", err)
	}

	return &result, nil
}

type ontologySubscriptionDeleteDTO struct {
	WebHookURL *string `json:"webHookURL,omitempty"`
	ID         *string `json:"id,omitempty"`
}

func (c *openBOSClient) deleteOntologySubscription(del ontologySubscriptionDeleteDTO) error {
	endpoint := fmt.Sprintf("%s/api/v1/ontology/full/version/subscribe", mockURL)

	if err := c.doMockRequest("DELETE", endpoint, nil, del, nil); err != nil {
		return fmt.Errorf("failed to delete ontology subscription: %v", err)
	}

	return nil
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
	Enums         map[string]string
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

	// dataTypeMap organizes datatypes in a map for simple lookup. Inside is a
	// slice to support complex data types.
	dataTypeMap := make(map[string][]ontologyDataTypeDTO)
	{
		// dataTypeComplexMap is an intermediate step to map dataTypes for lookup,
		// yet without unwrapping complex types.
		dataTypeComplexMap := make(map[string]ontologyDataTypeDTO)
		for _, dt := range ontology.DataTypes {
			dataTypeComplexMap[dt.ID] = dt
		}
		for _, dt := range ontology.DataTypes {
			dataTypeMap[dt.ID] = dt.unwrapComplexType(dataTypeComplexMap)
		}
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

		for _, datapointTemplate := range datapointTemplateMap[at.ID] {
			for _, dataType := range getDataTypes(datapointTemplate.TypeID, dataTypeMap) {
				// In case there is a complex datatype, it will be split into
				// multiple attributes.
				dataPoint := dataPoint{
					ID:            datapointTemplate.ID,
					Name:          datapointTemplate.Name,
					Tags:          datapointTemplate.Tags,
					TypeID:        datapointTemplate.TypeID,
					Min:           dataType.Min,
					Max:           dataType.Max,
					Enums:         dataType.Enums,
					Direction:     datapointTemplate.Direction,
					DisplayUnitID: getDisplayUnitID(dataType, unitMap),
				}
				assetTemplate.Datapoints = append(assetTemplate.Datapoints, dataPoint)
			}
		}

		for _, propertyTemplate := range propertyTemplateMap[at.ID] {
			for _, dataType := range getDataTypes(propertyTemplate.TypeID, dataTypeMap) {
				// In case there is a complex datatype, it will be split into
				// multiple attributes.
				property := property{
					ID:            propertyTemplate.ID,
					Name:          propertyTemplate.Name,
					Tags:          propertyTemplate.Tags,
					TypeID:        propertyTemplate.TypeID,
					Min:           dataType.Min,
					Max:           dataType.Max,
					Enums:         dataType.Enums,
					DisplayUnitID: getDisplayUnitID(dataType, unitMap),
				}
				assetTemplate.Properties = append(assetTemplate.Properties, property)
			}
		}

		assetTemplates = append(assetTemplates, assetTemplate)
	}

	return assetTemplates
}

func getDataTypes(typeID string, dataTypeMap map[string][]ontologyDataTypeDTO) []ontologyDataTypeDTO {
	dataTypes, exists := dataTypeMap[typeID]
	if !exists {
		log.Warn("client", "type %s not found", typeID)
		return nil
	}
	return dataTypes
}

func getDisplayUnitID(dataType ontologyDataTypeDTO, unitMap map[string]string) *string {
	var unitSymbol string
	if dataType.UnitID == "" {
		return &unitSymbol
	}
	unitSymbol, ok := unitMap[dataType.UnitID]
	if !ok {
		log.Warn("client", "unit %s not found", dataType.UnitID)
		return nil
	}
	return &unitSymbol
}

//  This file is part of the Eliona project.
//  Copyright © 2024 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package broker

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

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
	FormatVersion int32 `json:"version"` // TODO: Change this back to ontology version and use it, after ABB fixes the API implementation.
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
	Name   string                     `json:"name,omitempty"`
	Tags   []string                   `json:"tags,omitempty"`
	UnitID string                     `json:"unitId,omitempty"`
	Fields []ontologyDataTypeFieldDTO `json:"fields,omitempty"`
	Min    *float64                   `json:"min,omitempty"`
	Max    *float64                   `json:"max,omitempty"`
	Enums  map[string]string          `json:"enums,omitempty"`
}

type dataTypeUncomplexified struct {
	Format string
	Name   string
	UnitID string
	Min    *float64
	Max    *float64
	Enums  map[string]string
}

// unwrapComplexType recursively flattens complex datatypes into a slice of simple ones.
func (dt ontologyDataTypeDTO) unwrapComplexType(datatypeComplexMap map[string]ontologyDataTypeDTO, parentName string) []dataTypeUncomplexified {
	// If no fields, this is a primitive type; set its full path and return as a single-element slice.
	if len(dt.Fields) == 0 {
		dt.Name = parentName
		return []dataTypeUncomplexified{{
			Format: dt.Format,
			Name:   dt.Name,
			UnitID: dt.UnitID,
			Min:    dt.Min,
			Max:    dt.Max,
			Enums:  dt.Enums,
		}}
	}

	var result []dataTypeUncomplexified
	for _, childReference := range dt.Fields {
		child := datatypeComplexMap[childReference.TypeID]

		// Set child path by combining the parent's path with the current field's name.
		fieldPath := parentName
		fieldPath += "."
		fieldPath += childReference.Name

		// Recursively unwrap child
		result = append(result, child.unwrapComplexType(datatypeComplexMap, fieldPath)...)
	}

	return result
}

type ontologyDataTypeFieldDTO struct {
	Name   string `json:"name"` // Will never be empty
	TypeID string `json:"typeId"`
}

type ontologyDatapointTemplateDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name,omitempty"`
	AssetTemplateID string   `json:"assetTemplateId,omitempty"`
	SpaceTemplateID string   `json:"spaceTemplateId,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	TypeID          string   `json:"typeId,omitempty"`
	Direction       string   `json:"direction"`
	Perpetual       bool     `json:"perpetual"`
}

type ontologyPropertyTemplateDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name,omitempty"`
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

	datapoints []ontologyDatapointDTO
	properties []ontologyPropertyDTO
}

type ontologySpaceDTO struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	ParentID   string                  `json:"parentId,omitempty"`
	TemplateID string                  `json:"templateId"`
	Tags       []string                `json:"tags,omitempty"`
	Assets     []ontologySpaceAssetDTO `json:"assets,omitempty"`

	children []ontologySpaceDTO

	datapoints []ontologyDatapointDTO
	properties []ontologyPropertyDTO
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
	endpoint := "core/application/data"

	var ontology ontologyDTO
	if err := c.doRequest("GET", endpoint, nil, nil, &ontology); err != nil {
		return nil, err
	}

	return &ontology, nil
}

// getOntologyVersion retrieves the current version of the edge data.
func (c *openBOSClient) getOntologyVersion() (int32, error) {
	endpoint := "core/application/data/version"

	var version int32
	if err := c.doRequest("GET", endpoint, nil, nil, &version); err != nil {
		return 0, err
	}

	return version, nil
}

type subscriptionCreateDTO struct {
	MinSendTime       int32    `json:"minSendTime"`                // Minimum time between two events. To avoid events flushing. Highly recommended. If zero, events will be sent on the fly (NOT recommended). Min value: 1 mn
	MaxSendTime       int32    `json:"maxSendTime"`                // Maximum time between two events. Can be used to ensure the client application that the connection is alive. If nothing must be sent, the edge will send an empty event. Min value: 1 mn
	Timestamp         *string  `json:"timestamp,omitempty"`        // UTC date. To receive only datapoints or properties that change since the timestamp.
	WebHookURL        *string  `json:"webhookURL,omitempty"`       // URL of the webhook.
	WebHookRetries    int32    `json:"webhookRetries"`             // Interval of retries (in seconds) when an error occurs while sending an event.
	WebHookRetryDelay int32    `json:"webhookRetryDelay"`          // Number of retries when an error occurs while sending an event.
	WebHookLeaseTime  int32    `json:"webhookLeaseTime,omitempty"` // Life span of the webhook if the webhook connection is down. If not present, the webhook will never be destroyed. CAUTION server error 500 if too big (i.e. 60 minutes).
	WebhookPersist    *bool    `json:"webhookPersist,omitempty"`   // If true, the subscription will be kept alive when the edge restarts in the middle of the subscription. If false, the subscription is lost when the edge restarts.
	ContentType       *string  `json:"contentType,omitempty"`      // application/json for json (the default) or octet for base64.
	DesiredUnits      []string `json:"desiredUnits,omitempty"`     // List of units you want for certain datapoints.
}

type subscriptionResultDTO struct {
	ID         *string `json:"id,omitempty"`
	WebHookURL *string `json:"webhookURL,omitempty"`
}

func (c *openBOSClient) subscribeToOntologyChanges(configID int64) (*subscriptionResultDTO, error) {
	endpoint := "core/application/data/version/subscribe"

	webhookURL, err := url.JoinPath(c.webhookURL, fmt.Sprint(configID), "ontology-version")
	if err != nil {
		return nil, fmt.Errorf("joining URL for subscription: %v", err)
	}

	second := int32(1000)
	minute := 60 * second
	sub := subscriptionCreateDTO{
		MinSendTime:       5 * minute,
		WebHookURL:        common.Ptr(webhookURL),
		WebHookRetries:    3,
		WebHookRetryDelay: 5 * second,
		WebhookPersist:    common.Ptr(true),
	}

	var result subscriptionResultDTO
	if err := c.doRequest("POST", endpoint, nil, sub, &result); err != nil {
		return nil, fmt.Errorf("failed to subscribe to ontology changes: %v", err)
	}

	return &result, nil
}

type subscriptionDeleteDTO struct {
	WebHookURL *string `json:"webhookURL,omitempty"`
	ID         *string `json:"id,omitempty"`
}

func (c *openBOSClient) deleteOntologySubscription(del subscriptionDeleteDTO) error {
	endpoint := "core/application/data/version/subscribe"

	if err := c.doRequest("DELETE", endpoint, nil, del, nil); err != nil {
		return fmt.Errorf("failed to delete ontology subscription: %v", err)
	}

	return nil
}

func (c *openBOSClient) subscribeToDataChanges(configID int64) error {
	endpoint := "core/application/livedata/subscribe"

	webhookURL, err := url.JoinPath(c.webhookURL, fmt.Sprint(configID), "ontology-livedata")
	if err != nil {
		return fmt.Errorf("joining URL for subscription: %v", err)
	}

	second := int32(1000)
	minute := 60 * second
	sub := subscriptionCreateDTO{
		MinSendTime:       5 * minute,
		WebHookURL:        common.Ptr(webhookURL),
		WebHookRetries:    3,
		WebHookRetryDelay: 5 * second,
		WebHookLeaseTime:  5 * minute,
		WebhookPersist:    common.Ptr(true),
		ContentType:       common.Ptr("application/json"),
		// Info: There is also a parameter "desiredUnits" available. Implement if there is a use case.
	}

	if err := c.doRequest("POST", endpoint, nil, sub, nil); err != nil {
		return fmt.Errorf("failed to subscribe to data changes: %v", err)
	}

	// After successful subscription, trigger the initial synchronization
	refreshEndpoint := fmt.Sprintf("core/application/livedata/subscribe/refresh")

	req := struct {
		WebhookURL string `json:"webhookURL"`
	}{
		WebhookURL: webhookURL,
	}

	if err := c.doRequest("PUT", refreshEndpoint, nil, req, nil); err != nil {
		return fmt.Errorf("failed to trigger initial synchronization for webhookURL %s: %v", webhookURL, err)
	}

	return nil
}

func (c *openBOSClient) deleteDataSubscription(del subscriptionDeleteDTO) error {
	endpoint := "core/application/livedata/subscribe"

	if err := c.doRequest("DELETE", endpoint, nil, del, nil); err != nil {
		return fmt.Errorf("failed to delete data subscription: %v", err)
	}

	return nil
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

// getLiveAlarms retrieves live alarms since the given timestamp.
func (c *openBOSClient) getLiveAlarms(timestamp string) ([]ontologyFullLiveAlarmDTO, error) {
	endpoint := "core/application/livealarm"

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
	endpoint := "core/application/livealarm/ack"

	if err := c.doRequest("POST", endpoint, nil, ack, nil); err != nil {
		return err
	}

	return nil
}

type datapointTemplateInfo struct {
	ID         string
	Name       string
	Direction  string
	Attributes []templateAttributeInfo
}

type propertyTemplateInfo struct {
	ID         string
	Name       string
	Attributes []templateAttributeInfo
}

type templateAttributeInfo struct {
	Name          string
	DisplayUnitID *string
	Min           *float64
	Max           *float64
	Enums         map[string]string
}

type assetTemplate struct {
	ID         string
	Name       string
	Tags       []string
	Properties []propertyTemplateInfo
	Datapoints []datapointTemplateInfo
}

func (ontology ontologyDTO) getAssetTemplates(orphanDatapoints []ontologyDatapointDTO) []assetTemplate {
	datapointTemplateMap := make(map[string][]ontologyDatapointTemplateDTO)
	for _, dt := range ontology.DatapointTemplates {
		switch {
		case dt.AssetTemplateID != "" && dt.SpaceTemplateID != "":
			log.Warn("client", "datapoint template %s has both asset and space template ID", dt.ID)
			continue
		case dt.AssetTemplateID != "":
			datapointTemplateMap[dt.AssetTemplateID] = append(datapointTemplateMap[dt.AssetTemplateID], dt)
		case dt.SpaceTemplateID != "":
			datapointTemplateMap[dt.SpaceTemplateID] = append(datapointTemplateMap[dt.SpaceTemplateID], dt)
		default:
			log.Warn("client", "datapoint template %s has neither asset nor space template ID", dt.ID)
			continue
		}
	}

	// Root asset containing orphan datapoints
	rootAsset := ontologyAssetOrSpaceTemplateDTO{
		ID:   "root",
		Name: "root",
	}
	ontology.SpaceTemplates = append(ontology.SpaceTemplates, rootAsset)
	for _, orphanDatapoint := range orphanDatapoints {
		for _, dt := range ontology.DatapointTemplates {
			if orphanDatapoint.TemplateID == dt.ID {
				datapointTemplateMap["root"] = append(datapointTemplateMap["root"], dt)
			}
		}
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
	dataTypeMap := make(map[string][]dataTypeUncomplexified)
	{
		// dataTypeComplexMap is an intermediate step to map dataTypes for lookup,
		// yet without unwrapping complex types.
		dataTypeComplexMap := make(map[string]ontologyDataTypeDTO)
		for _, dt := range ontology.DataTypes {
			dataTypeComplexMap[dt.ID] = dt
		}
		for _, dt := range ontology.DataTypes {
			name := dt.Name
			if name == "" {
				// Name might be null, in that case let's use ID as a fallback.
				name = dt.ID
			}
			dataTypeMap[dt.ID] = dt.unwrapComplexType(dataTypeComplexMap, name)
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
			Datapoints: []datapointTemplateInfo{},
			Properties: []propertyTemplateInfo{},
		}

		for _, datapointTemplate := range datapointTemplateMap[at.ID] {
			dataPoint := datapointTemplateInfo{
				ID:        datapointTemplate.ID,
				Name:      datapointTemplate.Name,
				Direction: datapointTemplate.Direction,
			}
			for _, dataType := range getDataTypes(datapointTemplate.TypeID, dataTypeMap) {
				a := templateAttributeInfo{
					Name:          dataType.Name,
					Min:           dataType.Min,
					Max:           dataType.Max,
					Enums:         dataType.Enums,
					DisplayUnitID: getDisplayUnitID(dataType, unitMap),
				}
				dataPoint.Attributes = append(dataPoint.Attributes, a)
			}
			assetTemplate.Datapoints = append(assetTemplate.Datapoints, dataPoint)
		}

		for _, propertyTemplate := range propertyTemplateMap[at.ID] {
			property := propertyTemplateInfo{
				ID:   propertyTemplate.ID,
				Name: propertyTemplate.Name,
			}
			for _, dataType := range getDataTypes(propertyTemplate.TypeID, dataTypeMap) {
				a := templateAttributeInfo{
					Name:          dataType.Name,
					Min:           dataType.Min,
					Max:           dataType.Max,
					Enums:         dataType.Enums,
					DisplayUnitID: getDisplayUnitID(dataType, unitMap),
				}
				property.Attributes = append(property.Attributes, a)
			}
			assetTemplate.Properties = append(assetTemplate.Properties, property)
		}
		assetTemplates = append(assetTemplates, assetTemplate)
	}

	return assetTemplates
}

func getDataTypes(typeID string, dataTypeMap map[string][]dataTypeUncomplexified) []dataTypeUncomplexified {
	dataTypes, exists := dataTypeMap[typeID]
	if !exists {
		log.Warn("client", "type %s not found", typeID)
		return nil
	}
	return dataTypes
}

func getDisplayUnitID(dataType dataTypeUncomplexified, unitMap map[string]string) *string {
	var unitSymbol string
	if dataType.UnitID == "" {
		return nil
	}
	unitSymbol, ok := unitMap[dataType.UnitID]
	if !ok {
		log.Warn("client", "unit %s not found", dataType.UnitID)
		return nil
	}
	return &unitSymbol
}

func (c *openBOSClient) putData() error {
	endpoint := "ontology/datapointinstance/livedata"

	livedata := []struct {
		DataPointID string `json:"id"`
		Value       any    `json:"value"`
	}{} //todo
	// complex-encode
	var result []struct {
		DataPointID string `json:"id"`
		ErrorCode   string `json:"errorCode"`
		InnerError  string `json:"innerError"`
	}
	if err := c.doRequest("POST", endpoint, nil, livedata, &result); err != nil {
		return fmt.Errorf("failed to put data: %v", err)
	}

	log.Debug("client", "posting data: received %v results", len(result))
	var errs []error
	for _, r := range result {
		if r.ErrorCode != "" {
			errs = append(errs, fmt.Errorf("%+v", r))
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("received following error(s) while posting data: %v", errors.Join(errs...))
	}
	return nil
}

// subscribeToAlarmChanges subscribes to live alarm updates.
func (c *openBOSClient) subscribeToAlarmChanges(configID int64) error {
	endpoint := "core/application/livealarm/subscribe"

	webhookURL, err := url.JoinPath(c.webhookURL, fmt.Sprint(configID), "ontology-livealarm")
	if err != nil {
		return fmt.Errorf("joining URL for subscription: %v", err)
	}

	second := int32(1000)
	minute := 60 * second
	sub := subscriptionCreateDTO{
		MinSendTime:       5 * minute,
		WebHookURL:        common.Ptr(webhookURL),
		WebHookRetries:    3,
		WebHookRetryDelay: 5 * second,
		WebHookLeaseTime:  5 * minute,
		WebhookPersist:    common.Ptr(true),
		ContentType:       common.Ptr("application/json"),
	}

	if err := c.doRequest("POST", endpoint, nil, sub, nil); err != nil {
		return fmt.Errorf("failed to subscribe to alarm changes: %v", err)
	}

	return nil
}

// deleteAlarmSubscription deletes the live alarm subscription.
func (c *openBOSClient) deleteAlarmSubscription(del subscriptionDeleteDTO) error {
	endpoint := "core/application/livealarm/subscribe"

	if err := c.doRequest("DELETE", endpoint, nil, del, nil); err != nil {
		return fmt.Errorf("failed to delete alarm subscription: %v", err)
	}

	return nil
}

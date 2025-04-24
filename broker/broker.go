//  This file is part of the Eliona project.
//  Copyright © 2025 IoTEC AG. All Rights Reserved.
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
	appmodel "open-bos/app/model"
	"open-bos/complexdata"
	"open-bos/eliona"
	"strings"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const masterPropertyAttribute = "is_master"

var ErrNoUpdate = errors.New("no new version available")

func TestAuthentication(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}

	_, err = client.getOntologyVersion()
	return err
}

// [datapoint-attribution]
var datapointBelongingToAssetTemplate = make(map[string]datapointTemplatePreprocessedInfo) // map[datapointTemplateID]datapoints
type datapointTemplatePreprocessedInfo struct {
	name       string // datapoint name always
	subtype    string
	attributes []attributeTemplateInfo
}
type attributeTemplateInfo struct {
	name string // datatype (name or id) . uncomplexified path
}

func convertAssetTemplateToAssetType(template assetTemplate) api.AssetType {
	translatedName := "OpenBOS " + template.Name
	apiAsset := api.AssetType{
		Name: "open_bos_" + template.ID,
		Translation: *api.NewNullableTranslation(&api.Translation{
			En: &translatedName,
			De: &translatedName,
			Fr: &translatedName,
			It: &translatedName,
		}),
		Attributes: []api.AssetTypeAttribute{},
	}

	for _, dp := range template.Datapoints {
		subtype := determineSubtype(dp.Direction)
		var attributes []attributeTemplateInfo
		for _, dataType := range dp.DataTypes {
			mapping := convertMapping(dataType.Enums)

			// Only set nillables if they are not nil
			var min api.NullableFloat64
			if dataType.Min != nil {
				min = *api.NewNullableFloat64(dataType.Min)
			}
			var max api.NullableFloat64
			if dataType.Max != nil {
				max = *api.NewNullableFloat64(dataType.Max)
			}
			var unit api.NullableString
			if dataType.DisplayUnitID != nil {
				unit = *api.NewNullableString(dataType.DisplayUnitID)
			}

			attributeName := dp.Name
			if dataType.IsPartOfComplex {
				// Due to duplication, we cannot use the datapoint name. The dataType name allows this.
				// Append the sub-datatype names to the attribute name.
				parts := strings.SplitN(dataType.Name, ".", 2)
				if len(parts) > 1 {
					attributeName = dp.Name + "." + parts[1]
				} else {
					log.Error("broker", "should not happen: datatype %v is part of complex type, yet does not have more name parts", dataType.Name)
					attributeName = dp.Name
				}
			}

			attribute := api.AssetTypeAttribute{
				Name:    attributeName,
				Enable:  api.PtrBool(true),
				Subtype: subtype,
				Min:     min,
				Max:     max,
				Unit:    unit,
				Map:     mapping,
			}
			apiAsset.Attributes = append(apiAsset.Attributes, attribute)
			attributes = append(attributes, attributeTemplateInfo{
				name: attributeName,
			})
		}
		// [datapoint-attribution]
		datapointBelongingToAssetTemplate[dp.ID] = datapointTemplatePreprocessedInfo{
			name:       dp.Name,
			subtype:    string(subtype),
			attributes: attributes,
		}
	}

	// Properties are attributes that don't change often (our status subtype)
	for _, prop := range template.Properties {
		subtype := api.SUBTYPE_STATUS
		var attributes []attributeTemplateInfo
		for _, dataType := range prop.DataTypes {
			mapping := convertMapping(dataType.Enums)

			// Only set nillables if they are not nil
			var min api.NullableFloat64
			if dataType.Min != nil {
				min = *api.NewNullableFloat64(dataType.Min)
			}
			var max api.NullableFloat64
			if dataType.Max != nil {
				max = *api.NewNullableFloat64(dataType.Max)
			}
			var unit api.NullableString
			if dataType.DisplayUnitID != nil {
				unit = *api.NewNullableString(dataType.DisplayUnitID)
			}

			attributeName := prop.Name
			if dataType.IsPartOfComplex {
				// Due to duplication, we cannot use the datapoint name. The dataType name allows this.
				// Append the sub-datatype names to the attribute name.
				parts := strings.SplitN(dataType.Name, ".", 2)
				if len(parts) > 1 {
					attributeName = prop.Name + "." + parts[1]
				} else {
					log.Error("broker", "should not happen: datatype %v is part of complex type, yet does not have more name parts", dataType.Name)
					attributeName = prop.Name
				}
			}
			attribute := api.AssetTypeAttribute{
				Name:    attributeName,
				Enable:  api.PtrBool(true),
				Subtype: subtype,
				Min:     min,
				Max:     max,
				Unit:    unit,
				Map:     mapping,
			}
			apiAsset.Attributes = append(apiAsset.Attributes, attribute)
			attributes = append(attributes, attributeTemplateInfo{
				name: attributeName,
			})
		}
		// [datapoint-attribution]
		datapointBelongingToAssetTemplate[prop.ID] = datapointTemplatePreprocessedInfo{
			name:       prop.Name,
			subtype:    string(subtype),
			attributes: attributes,
		}
	}

	// TODO: Once APIv2 supports it, this should be a "Category"
	apiAsset.Attributes = append(apiAsset.Attributes, api.AssetTypeAttribute{
		Name:      masterPropertyAttribute,
		Subtype:   api.SUBTYPE_PROPERTY,
		Enable:    api.PtrBool(true),
		IsDigital: *api.NewNullableBool(api.PtrBool(true)),
		Map: []map[string]any{
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
	})
	return apiAsset
}

func determineSubtype(direction string) api.DataSubtype {
	switch strings.ToLower(direction) {
	case "feedback":
		return api.SUBTYPE_INPUT
	case "command", "commandandfeedback":
		return api.SUBTYPE_OUTPUT
	default:
		return api.SUBTYPE_INFO
	}
}

func convertMapping(enum map[string]string) []map[string]any {
	var mapping []map[string]any
	for key, value := range enum {
		mapping = append(mapping, map[string]any{
			"value": key,
			"map":   value,
		})
	}
	return mapping
}

func FetchOntology(config appmodel.Configuration) (ontologyVersion int32, assetTypes []api.AssetType, assets []eliona.Asset, err error) {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("creating instance of client: %v", err)
	}

	version, err := client.getOntologyVersion()
	if err != nil {
		return 0, nil, nil, fmt.Errorf("getting ontology version: %v", err)
	}
	if version == config.OntologyVersion {
		return 0, nil, nil, ErrNoUpdate
	}

	ontology, err := client.getOntology()
	if err != nil {
		return 0, nil, nil, fmt.Errorf("getting ontology: %v", err)
	}

	// [datapoint-attribution] Assign datapoints and properties to assets and spaces
	var orphanDatapoints []ontologyDatapointDTO
	{
		datapointsMap := make(map[string][]ontologyDatapointDTO)
		for _, datapointDTO := range ontology.Datapoints {
			if datapointDTO.AssetID != "" {
				datapointsMap[datapointDTO.AssetID] = append(datapointsMap[datapointDTO.AssetID], datapointDTO)
			}
			if datapointDTO.SpaceID != "" {
				datapointsMap[datapointDTO.SpaceID] = append(datapointsMap[datapointDTO.SpaceID], datapointDTO)
			}
			if datapointDTO.AssetID == "" && datapointDTO.SpaceID == "" {
				orphanDatapoints = append(orphanDatapoints, datapointDTO)
			}
		}
		propertiesMap := make(map[string][]ontologyPropertyDTO)
		for _, propertyDTO := range ontology.Properties {
			if propertyDTO.AssetID != "" {
				propertiesMap[propertyDTO.AssetID] = append(propertiesMap[propertyDTO.AssetID], propertyDTO)
			}
			if propertyDTO.SpaceID != "" {
				propertiesMap[propertyDTO.SpaceID] = append(propertiesMap[propertyDTO.SpaceID], propertyDTO)
			}
		}
		for i, asset := range ontology.Assets {
			ontology.Assets[i].datapoints = datapointsMap[asset.ID]
			ontology.Assets[i].properties = propertiesMap[asset.ID]
		}

		for i, space := range ontology.Spaces {
			ontology.Spaces[i].datapoints = datapointsMap[space.ID]
			ontology.Spaces[i].properties = propertiesMap[space.ID]
		}
	}

	ats := ontology.getAssetTemplates(orphanDatapoints)
	for _, assetTemplate := range ats {
		assetType := convertAssetTemplateToAssetType(assetTemplate)
		assetTypes = append(assetTypes, assetType)
	}

	spaces := make(map[string]*ontologySpaceDTO)
	// Build spaces map
	for _, space := range ontology.Spaces {
		spaceCopy := space
		spaces[space.ID] = &spaceCopy
	}

	var rootSpaces []ontologySpaceDTO
	// Build parent-child relationships
	for _, space := range ontology.Spaces {
		if parentSpace, exists := spaces[space.ParentID]; exists {
			parentSpace.children = append(parentSpace.children, *spaces[space.ID])
			// No need to reassign parentSpace back to the map since it's a pointer
		} else {
			// This is a Site, serving as a root
			rootSpaces = append(rootSpaces, space)
		}
	}

	// Build a map of assets for quick lookup
	assetsMap := make(map[string]ontologyAssetDTO)
	for _, asset := range ontology.Assets {
		assetsMap[asset.ID] = asset
	}

	for _, rootSpace := range rootSpaces {
		asset := eliona.Asset{
			ID:          rootSpace.ID,
			Name:        rootSpace.Name,
			TemplateID:  rootSpace.TemplateID,
			Config:      &config,
			IsRootSpace: true,
		}
		// Build the asset hierarchy based on spaces
		buildAssetHierarchy(&asset, &assets, spaces, assetsMap, config)
	}

	unassigned := eliona.Asset{
		ID:                  "",
		TemplateID:          "unassigned",
		Name:                "OpenBOS unassigned",
		Config:              &config,
		LocationalParentGAI: "",
		FunctionalParentGAI: "",
		IsRootSpace:         true,
	}
	assets = append(assets, unassigned)

	// Handle assets not associated with any space
	associatedAssetIDs := make(map[string]struct{})
	for _, space := range ontology.Spaces {
		for _, spaceAsset := range space.Assets {
			associatedAssetIDs[spaceAsset.ID] = struct{}{}
		}
	}
	for _, asset := range ontology.Assets {
		if _, associated := associatedAssetIDs[asset.ID]; !associated {
			// Asset not associated with any space; add to root
			assets = append(assets, eliona.Asset{
				ID:                  asset.ID,
				Name:                asset.Name,
				TemplateID:          asset.TemplateID,
				Config:              &config,
				LocationalParentGAI: unassigned.GetGAI(),
				FunctionalParentGAI: unassigned.GetGAI(),
			})
		}
	}

	assetTypes = append(assetTypes, api.AssetType{
		Name:   "open_bos_root",
		Custom: api.PtrBool(false),
		Translation: *api.NewNullableTranslation(&api.Translation{
			En: api.PtrString("OpenBOS root"),
			De: api.PtrString("OpenBOS root"),
		}),
		Attributes: []api.AssetTypeAttribute{
			{
				Name:    "status",
				Subtype: api.SUBTYPE_STATUS,
				Enable:  api.PtrBool(true),
				Translation: *api.NewNullableTranslation(&api.Translation{
					En: api.PtrString("Status"),
					De: api.PtrString("Status"),
				}),
				IsDigital: *api.NewNullableBool(api.PtrBool(true)),
				Min:       *api.NewNullableFloat64(api.PtrFloat64(0)),
				Max:       *api.NewNullableFloat64(api.PtrFloat64(2)),
				Map: []map[string]any{
					{
						"value": 0,
						"map":   "OK",
					},
					{
						"value": 1,
						"map":   "Error",
					},
					{
						"value": 2,
						"map":   "Fatal",
					},
				},
			},
		},
		AllowedInactivity: *api.NewNullableString(api.PtrString("00:05:00")),
	})
	assets = append(assets, eliona.Asset{
		ID:                  "root",
		TemplateID:          "root",
		Name:                "OpenBOS app",
		Config:              &config,
		LocationalParentGAI: "",
		FunctionalParentGAI: "",
		IsRootSpace:         true,
	})

	return ontology.Settings.OntologyVersion, assetTypes, assets, nil
}

func buildAssetHierarchy(asset *eliona.Asset, assets *[]eliona.Asset, spaces map[string]*ontologySpaceDTO, assetsMap map[string]ontologyAssetDTO, config appmodel.Configuration) {
	space, exists := spaces[asset.ID]
	if !exists {
		log.Error("broker", "Should not happen: space %s not found.", asset.ID)
		return
	}
	// [datapoint-attribution]
	// We need to merge attribute template information (name, subtype) with attribute instance information (instanceID, asset ID)
	var dps []appmodel.Datapoint
	for _, dp := range space.datapoints {
		datapoint, ok := datapointBelongingToAssetTemplate[dp.TemplateID]
		if !ok {
			log.Warn("broker", "datapoint not found for datapoint template %s", dp.TemplateID)
			continue
		}
		var attributes []appmodel.Attribute
		for _, attributeInfo := range datapoint.attributes {
			attributes = append(attributes, appmodel.Attribute{
				Name: attributeInfo.name,
			})
		}
		dps = append(dps, appmodel.Datapoint{
			Subtype:             datapoint.subtype,
			ProviderID:          dp.ID,
			AttributeNamePrefix: datapoint.name,
			Attributes:          attributes,
		})
	}
	for _, prop := range space.properties {
		datapoint, ok := datapointBelongingToAssetTemplate[prop.TemplateID]
		if !ok {
			log.Warn("broker", "datapoint not found for property template %s", prop.TemplateID)
			continue
		}
		var attributes []appmodel.Attribute
		for _, attributeInfo := range datapoint.attributes {
			attributes = append(attributes, appmodel.Attribute{
				Name: attributeInfo.name,
			})
		}
		dp := appmodel.Datapoint{
			Subtype:             datapoint.subtype,
			ProviderID:          prop.ID,
			AttributeNamePrefix: datapoint.name,
			Attributes:          attributes,
		}

		if prop.Value != nil {
			assetData := make(map[string]any)
			// Complex decode support
			if complexData, ok := prop.Value.(map[string]any); ok {
				decodedData := complexdata.DecodeComplexData(complexData, dp.AttributeNamePrefix)
				for k, v := range decodedData {
					assetData[k] = v
				}
			} else {
				// If not complex, find the attribute name and map directly
				if len(dp.Attributes) != 1 {
					log.Error("inconsistency", "received non-complex data %+v for property %v of datapoint %v, but found datapoint providerID %v with %v != 1 attributes", prop.Value, prop.ID, datapoint.name, dp.ProviderID, len(dp.Attributes))
					continue
				}
				assetData[dp.Attributes[0].Name] = prop.Value
			}
			dp.Data = assetData
		}

		dps = append(dps, dp)
	}
	asset.Datapoints = dps
	*assets = append(*assets, *asset)
	// Process child spaces
	for _, childSpace := range space.children {

		childAsset := eliona.Asset{
			ID:                  childSpace.ID,
			Name:                childSpace.Name,
			TemplateID:          childSpace.TemplateID,
			Config:              &config,
			LocationalParentGAI: asset.GetGAI(),
			FunctionalParentGAI: asset.GetGAI(),
		}
		if adheres, err := childAsset.AdheresToFilter(config.AssetFilter); err != nil {
			log.Error("broker", "checking if space adheres to filter: %v", err)
			continue
		} else if !adheres {
			log.Debug("broker", "skipped space ID %v name '%v' due to asset filter rule.", childSpace.ID, childSpace.Name)
			continue
		}
		buildAssetHierarchy(&childAsset, assets, spaces, assetsMap, config)
	}
	// Process assets associated with this space
	for _, spaceAsset := range space.Assets {
		assetDetails, exists := assetsMap[spaceAsset.ID]
		if !exists {
			log.Warn("broker", "asset %s in space %s not found. Skipping.", spaceAsset.ID, space.ID)
			continue // Asset not found; skip
		}

		// [datapoint-attribution]
		// We need to merge attribute template information (name, subtype) with attribute instance information (instanceID, asset ID)
		var dps []appmodel.Datapoint
		for _, dp := range assetDetails.datapoints {
			datapoint, ok := datapointBelongingToAssetTemplate[dp.TemplateID]
			if !ok {
				log.Warn("broker", "datapoint template not found for datapoint template %s", dp.TemplateID)
				continue
			}
			var attributes []appmodel.Attribute
			for _, attributeInfo := range datapoint.attributes {
				attributes = append(attributes, appmodel.Attribute{
					Name: attributeInfo.name,
				})
			}
			dps = append(dps, appmodel.Datapoint{
				Subtype:             datapoint.subtype,
				ProviderID:          dp.ID,
				AttributeNamePrefix: datapoint.name,
				Attributes:          attributes,
			})
		}
		for _, prop := range assetDetails.properties {
			datapoint, ok := datapointBelongingToAssetTemplate[prop.TemplateID]
			if !ok {
				log.Warn("broker", "datapoint template not found for property template %s", prop.TemplateID)
				continue
			}
			var attributes []appmodel.Attribute
			for _, attributeInfo := range datapoint.attributes {
				attributes = append(attributes, appmodel.Attribute{
					Name: attributeInfo.name,
				})
			}
			dp := appmodel.Datapoint{
				Subtype:             datapoint.subtype,
				ProviderID:          prop.ID,
				AttributeNamePrefix: datapoint.name,
				Attributes:          attributes,
			}

			if prop.Value != nil {
				assetData := make(map[string]any)
				// Complex decode support
				if complexData, ok := prop.Value.(map[string]any); ok {
					decodedData := complexdata.DecodeComplexData(complexData, dp.AttributeNamePrefix)
					for k, v := range decodedData {
						assetData[k] = v
					}
				} else {
					// If not complex, find the attribute name and map directly
					if len(dp.Attributes) != 1 {
						log.Error("inconsistency", "received non-complex data %+v for property %v of datapoint %v, but found datapoint providerID %v with %v != 1 attributes", prop.Value, prop.ID, assetDetails.ID, dp.ProviderID, len(dp.Attributes))
						continue
					}
					assetData[dp.Attributes[0].Name] = prop.Value
				}
				dp.Data = assetData
			}
			dps = append(dps, dp)
		}

		isMaster := int8(0)
		if spaceAsset.Master {
			isMaster = 1
		}
		assetInstance := eliona.Asset{
			ID:                  assetDetails.ID,
			Name:                assetDetails.Name,
			TemplateID:          assetDetails.TemplateID,
			Config:              &config,
			Datapoints:          dps,
			LocationalParentGAI: asset.GetGAI(),
			FunctionalParentGAI: asset.GetGAI(),

			IsMaster: isMaster,
		}
		if adheres, err := assetInstance.AdheresToFilter(config.AssetFilter); err != nil {
			log.Error("broker", "checking if asset adheres to filter: %v", err)
			continue
		} else if !adheres {
			log.Debug("broker", "skipped asset ID %v name '%v' due to asset filter rule.", assetInstance.ID, assetInstance.Name)
			continue
		}
		*assets = append(*assets, assetInstance)
	}
}

func SubscribeToOntologyChanges(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	if _, err := client.subscribeToOntologyChanges(config.Id); err != nil {
		return fmt.Errorf("subscribing: %v", err)
	}
	return nil
}

func SubscribeToDataChanges(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	if err := client.subscribeToDataChanges(config.Id); err != nil {
		return fmt.Errorf("subscribing: %v", err)
	}
	return nil
}

func FetchAlarmRules(config appmodel.Configuration) ([]AlarmRule, error) {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return nil, fmt.Errorf("creating instance of client: %v", err)
	}

	bosAlarms, err := client.getAlarmRules()
	if err != nil {
		return nil, fmt.Errorf("getting alarm rules: %v", err)
	}

	return bosAlarms, nil
}

func SubscribeToAlarms(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	if err := client.subscribeToAlarmChanges(config.Id); err != nil {
		return fmt.Errorf("subscribing: %v", err)
	}
	return nil
}

type AttributeData struct {
	Datapoint appmodel.Datapoint
	Value     any
}

func PutData(config appmodel.Configuration, attributesData []AttributeData) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	return client.putData(attributesData)
}

func AcknowledgeAlarm(config appmodel.Configuration, sessionID, ackedBy, comment string) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}

	ack := ontologyAlarmAckDTO{
		SessionID: sessionID,
		AckedBy:   ackedBy,
		Comment:   comment,
	}

	log.Debug("client", "Acknowledging alarm with session ID: %s", sessionID)

	if err := client.ackAlarm(ack); err != nil {
		log.Error("client", "Failed to acknowledge alarm: %v", err)
		return fmt.Errorf("acknowledging alarm: %v", err)
	}

	log.Debug("client", "Successfully acknowledged alarm with session ID: %s", sessionID)
	return nil
}

func CancelSubscriptions(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL, baseURL, tokenURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	if err := client.unsubscribeEverything(config.Id); err != nil {
		return fmt.Errorf("unsubscribing: %v", err)
	}
	return nil
}

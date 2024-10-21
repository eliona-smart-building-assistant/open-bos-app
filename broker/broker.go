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
	appmodel "open-bos/app/model"
	"open-bos/eliona"
	"strings"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const masterPropertyAttribute = "is_master"

var ErrNoUpdate = errors.New("no new version available")

// [datapoint-attribution]
var datapointTemplatesMap = make(map[string]datapointTemplateInfo) // map[datapointTemplateID]datapointTemplate
type datapointTemplateInfo struct {
	name    string
	subtype string
}

func convertAssetTemplateToAssetType(template assetTemplate) api.AssetType {
	translatedName := "OpenBOS " + template.Name
	apiAsset := api.AssetType{
		Name: "open_bos_" + template.ID,
		Translation: *api.NewNullableTranslation(&api.Translation{
			En: &translatedName,
		}),
		Attributes: []api.AssetTypeAttribute{},
	}

	for _, dp := range template.Datapoints {
		subtype := determineSubtype(dp.Direction)
		mapping := convertMapping(dp.Enums)
		attribute := api.AssetTypeAttribute{
			Name:    dp.Name,
			Subtype: subtype,
			Min:     *api.NewNullableFloat64(dp.Min),
			Max:     *api.NewNullableFloat64(dp.Max),
			Unit:    *api.NewNullableString(dp.DisplayUnitID),
			Map:     mapping,
		}
		apiAsset.Attributes = append(apiAsset.Attributes, attribute)
		// [datapoint-attribution]
		datapointTemplatesMap[dp.ID] = datapointTemplateInfo{
			name:    dp.Name,
			subtype: string(subtype),
		}
	}

	// Properties are attributes that don't change often (our status subtype)
	for _, prop := range template.Properties {
		subtype := api.SUBTYPE_STATUS
		mapping := convertMapping(prop.Enums)
		attribute := api.AssetTypeAttribute{
			Name:    prop.Name,
			Subtype: subtype,
			Min:     *api.NewNullableFloat64(prop.Min),
			Max:     *api.NewNullableFloat64(prop.Max),
			Unit:    *api.NewNullableString(prop.DisplayUnitID),
			Map:     mapping,
		}
		apiAsset.Attributes = append(apiAsset.Attributes, attribute)
		// [datapoint-attribution]
		datapointTemplatesMap[prop.ID] = datapointTemplateInfo{
			name:    prop.Name,
			subtype: string(subtype),
		}
	}

	// TODO: Once APIv2 supports it, this should be a "Category"
	apiAsset.Attributes = append(apiAsset.Attributes, api.AssetTypeAttribute{
		Name:      masterPropertyAttribute,
		Subtype:   api.SUBTYPE_PROPERTY,
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

func FetchOntology(config appmodel.Configuration) (ontologyVersion int32, assetTypes []api.AssetType, root eliona.Asset, err error) {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL)
	if err != nil {
		return 0, nil, eliona.Asset{}, fmt.Errorf("creating instance of client: %v", err)
	}

	if version, err := client.getOntologyVersion(); err != nil {
		return 0, nil, eliona.Asset{}, fmt.Errorf("getting ontology: %v", err)
	} else if version == config.OntologyVersion {
		return 0, nil, eliona.Asset{}, ErrNoUpdate
	}

	ontology, err := client.getOntology()
	if err != nil {
		return 0, nil, eliona.Asset{}, fmt.Errorf("getting ontology: %v", err)
	}

	ats := ontology.getAssetTemplates()
	ats = append(ats, assetTemplate{
		ID:   "root",
		Name: "OpenBOS root",
	})
	for _, assetTemplate := range ats {
		assetType := convertAssetTemplateToAssetType(assetTemplate)
		assetTypes = append(assetTypes, assetType)
	}

	root = eliona.Asset{
		ID:                    "",
		TemplateID:            "root",
		Name:                  "OpenBOS",
		Config:                &config,
		LocationalChildrenMap: make(map[string]eliona.Asset),
	}

	// [datapoint-attribution] Assign datapoints and properties to assets and spaces
	{
		datapointsMap := make(map[string][]ontologyDatapointDTO)
		for _, datapointDTO := range ontology.Datapoints {
			if datapointDTO.AssetID != "" {
				datapointsMap[datapointDTO.AssetID] = append(datapointsMap[datapointDTO.AssetID], datapointDTO)
			}
			if datapointDTO.SpaceID != "" {
				datapointsMap[datapointDTO.SpaceID] = append(datapointsMap[datapointDTO.SpaceID], datapointDTO)
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

	// Initialize spaces map with root space
	spaces := make(map[string]*ontologySpaceDTO)
	spaces[""] = &ontologySpaceDTO{}
	// Build spaces map
	for _, space := range ontology.Spaces {
		spaceCopy := space
		spaces[space.ID] = &spaceCopy
	}

	// Build parent-child relationships
	for _, space := range ontology.Spaces {
		if parentSpace, exists := spaces[space.ParentID]; exists {
			parentSpace.children = append(parentSpace.children, *spaces[space.ID])
			// No need to reassign parentSpace back to the map since it's a pointer
		}
	}

	// Build a map of assets for quick lookup
	assetsMap := make(map[string]ontologyAssetDTO)
	for _, asset := range ontology.Assets {
		assetsMap[asset.ID] = asset
	}

	// Build the asset hierarchy based on spaces
	buildAssetHierarchy(&root, spaces, assetsMap, config)

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
			root.FunctionalChildrenSlice = append(root.FunctionalChildrenSlice, eliona.Asset{
				ID:         asset.ID,
				Name:       asset.Name,
				TemplateID: asset.TemplateID,
				Config:     &config,
			})
		}
	}

	return ontology.Settings.Version, assetTypes, root, nil
}

func buildAssetHierarchy(asset *eliona.Asset, spaces map[string]*ontologySpaceDTO, assetsMap map[string]ontologyAssetDTO, config appmodel.Configuration) {
	space, exists := spaces[asset.ID]
	if !exists {
		log.Error("broker", "Should not happen: space %s not found.", asset.ID)
		return
	}
	// Process child spaces
	for _, childSpace := range space.children {

		// [datapoint-attribution]
		// We need to merge attribute template information (name, subtype) with attribute instance information (instanceID, asset ID)
		var attrs []appmodel.Attribute
		for _, dp := range childSpace.datapoints {
			datapointTemplateInfo, ok := datapointTemplatesMap[dp.TemplateID]
			if !ok {
				log.Warn("broker", "datapoint template not found for datapoint template %s", dp.TemplateID)
				continue
			}
			attrs = append(attrs, appmodel.Attribute{
				Name:       datapointTemplateInfo.name,
				Subtype:    datapointTemplateInfo.subtype,
				ProviderID: dp.ID,
			})
		}
		for _, prop := range childSpace.properties {
			datapointTemplateInfo, ok := datapointTemplatesMap[prop.TemplateID]
			if !ok {
				log.Warn("broker", "datapoint template not found for property template %s", prop.TemplateID)
				continue
			}
			attrs = append(attrs, appmodel.Attribute{
				Name:       datapointTemplateInfo.name,
				Subtype:    datapointTemplateInfo.subtype,
				ProviderID: prop.ID,
			})
		}

		childAsset := eliona.Asset{
			ID:                    childSpace.ID,
			Name:                  childSpace.Name,
			TemplateID:            childSpace.TemplateID,
			Config:                &config,
			LocationalChildrenMap: make(map[string]eliona.Asset),
			Attributes:            attrs,
		}
		buildAssetHierarchy(&childAsset, spaces, assetsMap, config)
		asset.LocationalChildrenMap[childSpace.ID] = childAsset
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
		var attrs []appmodel.Attribute
		for _, dp := range assetDetails.datapoints {
			datapointTemplateInfo, ok := datapointTemplatesMap[dp.TemplateID]
			if !ok {
				log.Warn("broker", "datapoint template not found for datapoint template %s", dp.TemplateID)
				continue
			}
			attrs = append(attrs, appmodel.Attribute{
				Name:       datapointTemplateInfo.name,
				Subtype:    datapointTemplateInfo.subtype,
				ProviderID: dp.ID,
			})
		}
		for _, prop := range assetDetails.properties {
			datapointTemplateInfo, ok := datapointTemplatesMap[prop.TemplateID]
			if !ok {
				log.Warn("broker", "datapoint template not found for property template %s", prop.TemplateID)
				continue
			}
			attrs = append(attrs, appmodel.Attribute{
				Name:       datapointTemplateInfo.name,
				Subtype:    datapointTemplateInfo.subtype,
				ProviderID: prop.ID,
			})
		}

		isMaster := int8(0)
		if spaceAsset.Master {
			isMaster = 1
		}
		assetInstance := eliona.Asset{
			ID:         assetDetails.ID,
			Name:       assetDetails.Name,
			TemplateID: assetDetails.TemplateID,
			Config:     &config,

			IsMaster: isMaster,
		}
		asset.LocationalChildrenMap[spaceAsset.ID] = assetInstance
	}
}

func SubscribeToOntologyChanges(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	if _, err := client.subscribeToOntologyChanges(config.Id); err != nil {
		return fmt.Errorf("subscribing: %v", err)
	}
	return nil
}

func SubscribeToDataChanges(config appmodel.Configuration) error {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret, config.AppPublicAPIURL)
	if err != nil {
		return fmt.Errorf("creating instance of client: %v", err)
	}
	if _, err := client.subscribeToDataChanges(config.Id); err != nil {
		return fmt.Errorf("subscribing: %v", err)
	}
	return nil
}

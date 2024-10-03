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
	"fmt"
	appmodel "open-bos/app/model"
	"open-bos/eliona"
	"strings"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
)

func convertAssetTemplateToAssetType(template assetTemplate) api.AssetType {
	apiAsset := api.AssetType{
		Name: "openBOS-" + template.ID,
		Translation: *api.NewNullableTranslation(&api.Translation{
			En: &template.Name,
		}),
		Attributes: []api.AssetTypeAttribute{},
	}

	for _, dp := range template.Datapoints {
		subtype := determineSubtype(dp.Direction)
		attribute := api.AssetTypeAttribute{
			Name:    dp.Name,
			Subtype: subtype,
			Unit:    *api.NewNullableString(dp.DisplayUnitID),
		}
		apiAsset.Attributes = append(apiAsset.Attributes, attribute)
	}

	// Properties are attributes that don't change often (our status subtype)
	for _, prop := range template.Properties {
		attribute := api.AssetTypeAttribute{
			Name:    prop.Name,
			Subtype: api.SUBTYPE_STATUS,
			Unit:    *api.NewNullableString(prop.DisplayUnitID),
		}
		apiAsset.Attributes = append(apiAsset.Attributes, attribute)
	}

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

func FetchAssets(config appmodel.Configuration) ([]api.AssetType, eliona.Asset, error) {
	client, err := newOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret)
	if err != nil {
		return nil, eliona.Asset{}, fmt.Errorf("creating instance of client: %v", err)
	}
	ontology, err := client.getOntology()
	if err != nil {
		return nil, eliona.Asset{}, fmt.Errorf("getting ontology: %v", err)
	}

	ats := ontology.getAssetTemplates()
	ats = append(ats, assetTemplate{
		ID:   "0",
		Name: "OpenBOS root",
	})
	var assetTypes []api.AssetType
	for _, assetTemplate := range ats {
		assetType := convertAssetTemplateToAssetType(assetTemplate)
		assetTypes = append(assetTypes, assetType)
	}
	root := eliona.Asset{
		ID:         "0",
		TemplateID: "0",
		Name:       "Root",
		Config:     &config,
	}
	for _, asset := range ontology.Assets {
		root.DevicesSlice = append(root.DevicesSlice, eliona.Asset{
			ID:         asset.ID,
			Name:       asset.Name,
			TemplateID: asset.TemplateID,
			Config:     &config,
		})
	}
	return assetTypes, root, nil
}

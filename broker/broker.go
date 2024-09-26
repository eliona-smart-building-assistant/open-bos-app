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
	"strings"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
)

func GetAssetTypes(config appmodel.Configuration) ([]api.AssetType, error) {
	client, err := NewOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("creating instance of client: %v", err)
	}
	assetTemplates, err := client.getAssetTemplates()
	if err != nil {
		return nil, fmt.Errorf("getting functional block template: %v", err)
	}
	var assetTypes []api.AssetType
	for _, assetTemplate := range assetTemplates {
		assetType := convertAssetTemplateToAssetType(assetTemplate)
		assetTypes = append(assetTypes, assetType)
	}
	return assetTypes, nil
}

func convertAssetTemplateToAssetType(template AssetTemplate) api.AssetType {
	// Initialize AssetType
	apiAsset := api.AssetType{
		Name:       template.Id, // Set the AssetType name to the template's Id
		Attributes: []api.AssetTypeAttribute{},
	}

	// Add DataPoints as Attributes with relevant Subtype
	for _, dp := range template.Datapoints {
		subtype := determineSubtype(dp.Direction)
		attribute := api.AssetTypeAttribute{
			Name:    dp.Name,
			Subtype: subtype,
			Unit:    *api.NewNullableString(dp.DisplayUnitId),
		}
		apiAsset.Attributes = append(apiAsset.Attributes, attribute)
	}

	// Add Properties as Attributes with Subtype Status
	for _, prop := range template.Properties {
		attribute := api.AssetTypeAttribute{
			Name:    prop.Name,
			Subtype: api.SUBTYPE_STATUS,
			Unit:    *api.NewNullableString(prop.DisplayUnitId),
		}
		apiAsset.Attributes = append(apiAsset.Attributes, attribute)
	}

	return apiAsset
}

// Helper function to determine the subtype for DataPoints
func determineSubtype(direction string) api.DataSubtype {
	switch strings.ToLower(direction) {
	case "feedback":
		return api.SUBTYPE_INPUT
	case "control", "feedbackandcontrol":
		return api.SUBTYPE_OUTPUT
	default:
		return api.SUBTYPE_INFO
	}
}

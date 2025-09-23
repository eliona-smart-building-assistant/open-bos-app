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

package eliona

import (
	"fmt"
	"net/http"
	appmodel "open-bos/app/model"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v3"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/client"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func CreateAssets(config appmodel.Configuration, assets []Asset) error {
	var elionaAssets []asset.AssetLikeWithParentReferences
	for _, a := range assets {
		elionaAssets = append(elionaAssets, asset.AssetLikeWithParentReferences(&a))
	}

	roots, err := fetchRootsFromEliona(config.ApiKey, assets)
	if err != nil {
		return fmt.Errorf("fetching root: %v", err)
	}
	log.Debug("eliona", "started creating assets for site %s", config.ElionaSiteId)
	// todo: this does not return assets created anymore, but total number of assets!
	assetsCreated, err := asset.CreateAssetsBulk(client.ApiEndpointString(), config.ApiKey, elionaAssets)
	if err != nil {
		return err
	}

	for _, root := range roots {
		// Restore root to allow moving the whole structure to subdirectories in Eliona.
		_, err := asset.UpsertAsset(client.ApiEndpointString(), config.ApiKey, *root)
		if err != nil {
			return fmt.Errorf("returning root asset to its original state")
		}
	}
	log.Debug("eliona", "finished creating %v assets", assetsCreated)
	if assetsCreated != 0 {
		if err := notifyUser(config.ApiKey, config.UserId, assetsCreated); err != nil {
			return fmt.Errorf("notifying user about CAC: %v", err)
		}
	}

	log.Debug("eliona", "started upserting properties data for assets")
	if err := upsertData(client.ApiEndpointString(), config.ApiKey, assets); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	log.Debug("eliona", "finished upserting properties data for assets")

	return nil
}

func fetchRootsFromEliona(apiKey string, assets []Asset) ([]*api.Asset, error) {
	var apiRoots []*api.Asset
	for _, asset := range assets {
		if asset.IsRootSpace {
			rootID, err := asset.GetAssetID()
			if err != nil {
				return nil, fmt.Errorf("getting root asset ID: %v", err)
			}
			if rootID == nil {
				return nil, nil
			}
			root, err := getAsset(apiKey, *rootID)
			if err != nil {
				return nil, fmt.Errorf("getting root asset from API: %v", err)
			}
			if root == nil {
				continue
			}

			apiRoots = append(apiRoots, root)
		}
	}

	return apiRoots, nil
}

func getAsset(apiKey string, assetId int32) (*api.Asset, error) {
	a, res, err := client.NewClient(client.ApiEndpointString()).AssetsAPI.
		GetAssetById(client.AuthenticationContext(apiKey), assetId).
		Execute()
	if res != nil && res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return a, err
}

func upsertData(apiEndpoint string, apiKey string, assets []Asset) error {
	for _, a := range assets {
		assetID, err := a.GetAssetID()
		if err != nil {
			return fmt.Errorf("getting asset ID: %v", err)
		}
		if assetID == nil {
			return fmt.Errorf("assetID is nil for asset %v", a.GetGAI())
		}

		for _, datapoint := range a.Datapoints {
			if datapoint.Data == nil || len(datapoint.Data) == 0 {
				continue
			}
			if err := UpsertAssetData(apiEndpoint, apiKey, *assetID, datapoint.Data, time.Now(), api.DataSubtype(datapoint.Subtype)); err != nil {
				return fmt.Errorf("upserting asset data %v for asset ID %v subtype %v: %v", datapoint.Data, *assetID, datapoint.Subtype, err)
			}
		}
	}
	return nil
}

func notifyUser(apiKey string, userId string, assetsCreated int) error {
	receipt, _, err := client.NewClient(client.ApiEndpointString()).CommunicationAPI.
		PostNotification(client.AuthenticationContext(apiKey)).
		Notification(
			api.Notification{
				User: userId,
				Message: *api.NewNullableTranslation(&api.Translation{
					De: api.PtrString(fmt.Sprintf("OpenBOS-Ontologie wurde synchronisiert. %v Assets werden synchron gehalten.", assetsCreated)),
					En: api.PtrString(fmt.Sprintf("OpenBOS ontology was synchronized. %v assets are kept in sync.", assetsCreated)),
				}),
			}).
		Execute()
	log.Debug("eliona", "posted notification about CAC: %v", receipt)
	if err != nil {
		return fmt.Errorf("posting CAC notification: %v", err)
	}
	return nil
}

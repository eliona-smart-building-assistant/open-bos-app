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

package eliona

import (
	"fmt"
	appmodel "open-bos/app/model"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func CreateAssets(config appmodel.Configuration, root Asset) error {
	for _, projectId := range config.ProjectIDs {
		assetsCreated, err := asset.CreateAssets(asset.Root(&root), projectId)
		if err != nil {
			return err
		}
		if assetsCreated != 0 {
			if err := notifyUser(config.UserId, projectId, assetsCreated); err != nil {
				return fmt.Errorf("notifying user about CAC: %v", err)
			}
		}
		if err := upsertDataRecursively(root, projectId); err != nil {
			return fmt.Errorf("upserting data: %v", err)
		}
	}
	return nil
}

func upsertDataRecursively(node Asset, projectId string) error {
	assetID, err := node.GetAssetID(projectId)
	if err != nil {
		return fmt.Errorf("getting asset ID: %v", err)
	}
	if assetID == nil {
		return fmt.Errorf("assetID is nil for asset %v, project %v", node.GetGAI(), projectId)
	}

	for _, datapoint := range node.Datapoints {
		if datapoint.Data == nil || len(datapoint.Data) == 0 {
			continue
		}
		if err := UpsertAssetData(*assetID, datapoint.Data, time.Now(), api.DataSubtype(datapoint.Subtype)); err != nil {
			return fmt.Errorf("upserting asset data %v for asset ID %v subtype %v: %v", datapoint.Data, *assetID, datapoint.Subtype, err)
		}
	}

	for _, child := range node.getLocationalAssetChildren() {
		if err := upsertDataRecursively(child, projectId); err != nil {
			return err
		}
	}

	for _, child := range node.getFunctionalAssetChildren() {
		if err := upsertDataRecursively(child, projectId); err != nil {
			return err
		}
	}

	return nil
}

func notifyUser(userId string, projectId string, assetsCreated int) error {
	receipt, _, err := client.NewClient().CommunicationAPI.
		PostNotification(client.AuthenticationContext()).
		Notification(
			api.Notification{
				User:      userId,
				ProjectId: *api.NewNullableString(&projectId),
				Message: *api.NewNullableTranslation(&api.Translation{
					De: api.PtrString(fmt.Sprintf("OpenBOS App hat %d neue Assets angelegt. Diese sind nun im Asset-Management verfügbar.", assetsCreated)),
					En: api.PtrString(fmt.Sprintf("OpenBOS app added %v new assets. They are now available in Asset Management.", assetsCreated)),
				}),
			}).
		Execute()
	log.Debug("eliona", "posted notification about CAC: %v", receipt)
	if err != nil {
		return fmt.Errorf("posting CAC notification: %v", err)
	}
	return nil
}

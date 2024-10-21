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
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const ClientReference string = "open-bos"

func UpsertAssetData(assetID int32, assetData map[string]any, timestamp time.Time, subtype api.DataSubtype) error {
	cr := ClientReference
	log.Debug("Eliona", "upserting data for asset '%v'", assetID)

	data := api.Data{
		AssetId:         assetID,
		Subtype:         subtype,
		Timestamp:       *api.NewNullableTime(&timestamp),
		Data:            assetData,
		ClientReference: *api.NewNullableString(&cr),
		// AssetTypeName: api.NullableString{}, No need to fill, it's only for seleciton
	}
	if err := asset.UpsertDataIfAssetExists(data); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	return nil
}

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
	"context"
	"fmt"
	appmodel "open-bos/app/model"
	conf "open-bos/db/helper"

	"github.com/eliona-smart-building-assistant/go-eliona/asset"
)

type Asset struct {
	ID   string
	Name string

	TemplateID string

	LocationsMap map[string]Asset
	DevicesSlice []Asset

	Config *appmodel.Configuration
}

func (d *Asset) GetName() string {
	return d.Name
}

func (d *Asset) GetDescription() string {
	return ""
}

func (d *Asset) GetAssetType() string {
	return "openBOS-" + d.TemplateID
}

func (d *Asset) GetGAI() string {
	return "openBOS-" + d.ID
}

func (d *Asset) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *Asset) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *d.Config, projectID, d.GetGAI(), assetID, d.ID); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (r *Asset) GetLocationalChildren() []asset.LocationalNode {
	locationalChildren := make([]asset.LocationalNode, 0, len(r.LocationsMap))
	for _, room := range r.LocationsMap {
		roomCopy := room // Create a copy of room
		locationalChildren = append(locationalChildren, &roomCopy)
	}
	return locationalChildren
}

func (r *Asset) GetFunctionalChildren() []asset.FunctionalNode {
	functionalChildren := make([]asset.FunctionalNode, 0, len(r.DevicesSlice))
	for i := range r.DevicesSlice {
		functionalChildren = append(functionalChildren, &r.DevicesSlice[i])
	}
	return functionalChildren
}

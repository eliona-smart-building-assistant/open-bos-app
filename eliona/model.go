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
	"context"
	"fmt"
	appmodel "open-bos/app/model"
	conf "open-bos/db/helper"

	"github.com/eliona-smart-building-assistant/go-eliona/utils"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

type Asset struct {
	ID   string `eliona:"id,filterable"`
	Name string `eliona:"name,filterable"`

	TemplateID string `eliona:"templateID,filterable"`

	IsMaster int8 `eliona:"is_master" subtype:"property"`

	IsRootSpace bool // Used for sites that are roots of the structure (Only these are allowed to be moved within Eliona asset structure)

	LocationalParentGAI string
	FunctionalParentGAI string

	Datapoints []appmodel.Datapoint

	Config *appmodel.Configuration
}

func (d *Asset) GetName() string {
	return d.Name
}

func (d *Asset) AdheresToFilter(filter [][]appmodel.FilterRule) (bool, error) {
	f := appFilterToCommonFilter(filter)
	fp, err := utils.StructToMap(d)
	if err != nil {
		return false, fmt.Errorf("converting struct to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func (d *Asset) GetDescription() string {
	return ""
}

func (d *Asset) GetAssetType() string {
	return "open_bos_" + d.TemplateID
}

func (d *Asset) GetGAI() string {
	return "open_bos_" + d.ID
}

func (d *Asset) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *Asset) SetAssetID(elionaAssetID int32, projectID string) error {
	ctx := context.Background()
	assetID, err := conf.InsertAsset(ctx, *d.Config, projectID, d.GetGAI(), elionaAssetID, d.ID, d.IsRootSpace)
	if err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}

	if err := conf.InsertAssetAttributes(ctx, assetID, d.Datapoints); err != nil {
		return fmt.Errorf("inserting asset subtypes to config db: %v", err)
	}

	return nil
}

func (a *Asset) GetLocationalParentGAI() string {
	return a.LocationalParentGAI
}

func (a *Asset) GetFunctionalParentGAI() string {
	return a.FunctionalParentGAI
}

func appFilterToCommonFilter(input [][]appmodel.FilterRule) [][]common.FilterRule {
	result := make([][]common.FilterRule, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = make([]common.FilterRule, len(input[i]))
		for j := 0; j < len(input[i]); j++ {
			result[i][j] = common.FilterRule{
				Parameter: input[i][j].Parameter,
				Regex:     input[i][j].Regex,
			}
		}
	}
	return result
}

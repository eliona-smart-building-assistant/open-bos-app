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

package app

import (
	"context"
	"errors"
	"net/http"
	apiserver "open-bos/api/generated"
	apiservices "open-bos/api/services"
	appmodel "open-bos/app/model"
	"open-bos/broker"
	dbhelper "open-bos/db/helper"
	"open-bos/eliona"
	"sync"
	"time"

	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func Initialize() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// Init the app before the first run.
	app.Init(conn, app.AppName(),
		app.ExecSqlFile("db/init.sql"),
	)
}

var notifyNoConfigsOnce sync.Once

func CollectData() {
	configs, err := dbhelper.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("dbhelper", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		notifyNoConfigsOnce.Do(func() {
			log.Info("dbhelper", "No configs in DB. Please configure the app in Eliona.")
		})
		return
	}

	for _, config := range configs {
		if !config.Enable {
			if config.Active {
				dbhelper.SetConfigActiveState(context.Background(), config, false)
			}
			continue
		}

		if !config.Active {
			dbhelper.SetConfigActiveState(context.Background(), config, true)
			log.Info("dbhelper", "Collecting initialized with Configuration %d:\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Project IDs: %v\n",
				config.Id,
				config.Enable,
				config.RefreshInterval,
				config.RequestTimeout,
				config.ProjectIDs)
		}

		common.RunOnceWithParam(func(config appmodel.Configuration) {
			log.Info("main", "Collecting %d started.", config.Id)
			if err := collectResources(&config); err != nil {
				return // Error is handled in the method itself.
			}
			log.Info("main", "Collecting %d finished.", config.Id)

			time.Sleep(time.Hour * time.Duration(config.RefreshInterval))
		}, config, config.Id)
	}
}

func CollectConfigData(configID int64) {
	config, err := dbhelper.GetConfig(context.Background(), configID)
	if err != nil {
		log.Error("dbhelper", "Couldn't read config %d from DB: %v", configID, err)
		return
	}

	if !config.Enable {
		if config.Active {
			dbhelper.SetConfigActiveState(context.Background(), config, false)
		}
		return
	}
	if !config.Active {
		dbhelper.SetConfigActiveState(context.Background(), config, true)
	}

	log.Info("main", "Collecting %d triggered by update.", config.Id)
	if err := collectResources(&config); err != nil {
		return // Error is handled in the method itself.
	}
	log.Info("main", "Collecting %d finished.", config.Id)
}

func collectResources(config *appmodel.Configuration) error {
	version, assetTypes, root, err := broker.FetchOntology(*config)
	if errors.Is(err, broker.ErrNoUpdate) {
		log.Debug("broker", "ontology is up-to-date")
		return nil
	}
	if err != nil {
		log.Error("broker", "fetching assets: %v", err)
		return err
	}
	for _, assetType := range assetTypes {
		if err := asset.InitAssetType(assetType)(nil); err != nil {
			log.Error("eliona", "initializing asset type: %v", err)
			return err
		}
	}
	if err := eliona.CreateAssets(*config, &root); err != nil {
		log.Error("eliona", "creating assets: %v", err)
		return err
	}

	config.OntologyVersion = version
	dbhelper.UpdateConfigOntologyVersion(context.Background(), *config)

	return nil
}

// ListenForOutputChanges listens to output attribute changes from Eliona. Delete if not needed.
func ListenForOutputChanges() {
	for { // We want to restart listening in case something breaks.
		outputs, err := eliona.ListenForOutputChanges()
		if err != nil {
			log.Error("eliona", "listening for output changes: %v", err)
			return
		}
		for output := range outputs {
			if cr := output.ClientReference.Get(); cr != nil && *cr == eliona.ClientReference {
				// Just an echoed value this app sent.
				continue
			}
			asset, err := dbhelper.GetAssetById(output.AssetId)
			if err != nil {
				log.Error("dbhelper", "getting asset by assetID %v: %v", output.AssetId, err)
				return
			}
			if err := outputData(asset, output.Data); err != nil {
				log.Error("dbhelper", "outputting data (%v) for config %v and assetId %v: %v", output.Data, asset.Config.Id, asset.AssetID, err)
				return
			}
		}
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
}

// outputData implements passing output data to broker. Remove if not needed.
func outputData(asset appmodel.Asset, data map[string]interface{}) error {
	// Do the output magic here.
	return nil
}

// ListenApi starts the API server and listen for requests
func ListenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"),
		frontend.NewEnvironmentHandler(
			utilshttp.NewCORSEnabledHandler(
				apiserver.NewRouter(
					apiserver.NewConfigurationAPIController(apiservices.NewConfigurationAPIService()),
					apiserver.NewVersionAPIController(apiservices.NewVersionAPIService()),
					apiserver.NewCustomizationAPIController(apiservices.NewCustomizationAPIService()),
				))))
	log.Fatal("main", "API server: %v", err)
}

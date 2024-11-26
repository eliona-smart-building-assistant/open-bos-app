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
	"fmt"
	"net/http"
	apiserver "open-bos/api/generated"
	apiservices "open-bos/api/services"
	appmodel "open-bos/app/model"
	"open-bos/broker"
	dbhelper "open-bos/db/helper"
	"open-bos/eliona"
	"strings"
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

			if err := broker.SubscribeToOntologyChanges(config); err != nil {
				log.Error("broker", "subscribing to ontology changes: %v", err)
				return
			}
			log.Info("main", "Subscribed to ontology updates of config %d", config.Id)

			if err := broker.SubscribeToDataChanges(config); err != nil {
				log.Error("broker", "subscribing to data changes: %v", err)
				return
			}
			log.Info("main", "Subscribed to data updates of config %d", config.Id)

			if err := broker.SubscribeToAlarms(config); err != nil {
				log.Error("broker", "subscribing to alarm changes: %v", err)
				return
			}
			log.Info("main", "Subscribed to alarm updates of config %d", config.Id)

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

type AttributeDataUpdate struct {
	ConfigID            int64
	DatapointProviderID string
	Timestamp           time.Time
	Value               any
}

// TODO: change to bulk upsert
func UpdateDataPointInEliona(update AttributeDataUpdate) {
	config, err := dbhelper.GetConfig(context.Background(), update.ConfigID)
	if err != nil {
		log.Error("dbhelper", "Couldn't read config %d from DB: %v", update.ConfigID, err)
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

	assetData := make(map[string]any)
	datapoint, err := dbhelper.GetDatapointById(update.DatapointProviderID, config.Id)
	if err != nil {
		log.Error("dbhelper", "getting datapoint by ID %v for config %v: %v", update.DatapointProviderID, config.Id, err)
		return
	}
	// Complex decode support
	if complexData, ok := update.Value.(map[string]any); ok {
		decodedData := decodeComplexData(complexData, datapoint.AttributeNamePrefix)
		for k, v := range decodedData {
			assetData[k] = v
		}
	} else {
		// If not complex, find the attribute name and map directly
		if len(datapoint.Attributes) != 1 {
			log.Error("inconsistency", "received non-complex data %+v, but found datapoint providerID %v with %v != 1 attributes", update, datapoint.ProviderID, len(datapoint.Attributes))
			return
		}
		assetData[datapoint.Attributes[0].Name] = update.Value
	}

	// TODO: Upsert data
}

func decodeComplexData(value map[string]any, parentPath string) map[string]any {
	flattened := make(map[string]any)
	for key, val := range value {
		currentPath := parentPath
		if currentPath != "" {
			currentPath += "." + key
		} else {
			currentPath = key
		}

		// Handle nested complex values recursively
		if nested, ok := val.(map[string]any); ok {
			nestedData := decodeComplexData(nested, currentPath)
			for nestedKey, nestedVal := range nestedData {
				flattened[nestedKey] = nestedVal
			}
		} else {
			flattened[currentPath] = val
		}
	}
	return flattened
}

type AlarmUpdate struct {
	ConfigID            int64
	DatapointInstanceId string
	Timestamp           time.Time
	AlarmID             string
	Name                string
	Description         string
	Trigger             string
	Active              bool
	Acked               bool
	Closed              bool
	TimeStamp           string
	Quality             string
	Value               any
	AckedBy             string
	Comment             string
	NeedAcknowledge     bool
	Severity            string
	AssetId             string
	SpaceId             string
	AssetName           string
	SpaceName           string
	DatapointName       string
	UnitSymbol          string
	Tags                []string
}

func (alarm AlarmUpdate) getPriority() int {
	switch alarm.Severity {
	case "Critical", "Urgent":
		return 1 // High priority
	case "High":
		return 2 // Medium priority
	case "Low":
		return 3 // Low priority
	case "Log":
		return 10 // Info
	default:
		return 10 // Default to lowest priority if severity is unknown
	}
}

// buildAlarmMessage builds alarm message to Eliona format.
func (alarm AlarmUpdate) buildAlarmMessage() map[string]interface{} {
	message := make(map[string]interface{})

	languageCodes := []string{"de", "en", "fr", "it"}

	var descriptionPart string
	if alarm.Description != "" {
		descriptionPart = fmt.Sprintf(": %s", alarm.Description)
	}
	templateCome := fmt.Sprintf("%s%s {{asset.name}} ({{alarm.val}})", alarm.Name, descriptionPart)

	come := make(map[string]string)
	for _, lang := range languageCodes {
		come[lang] = templateCome
	}

	goneTranslations := map[string]string{
		"de": fmt.Sprintf("%s behoben", alarm.Name),
		"en": fmt.Sprintf("%s resolved", alarm.Name),
		"fr": fmt.Sprintf("%s résolu", alarm.Name),
		"it": fmt.Sprintf("%s risolto", alarm.Name),
	}

	gone := make(map[string]string)
	for _, lang := range languageCodes {
		gone[lang] = goneTranslations[lang]
	}

	message["come"] = come
	message["gone"] = gone

	return message
}

func (alarm AlarmUpdate) getAckMessage() string {
	return fmt.Sprintf("%s: %s", alarm.AckedBy, alarm.Comment)
}

func UpdateAlarmInEliona(update AlarmUpdate) {
	config, err := dbhelper.GetConfig(context.Background(), update.ConfigID)
	if err != nil {
		log.Error("dbhelper", "Couldn't read config %d from DB: %v", update.ConfigID, err)
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
	datapoint, err := dbhelper.GetDatapointById(update.DatapointInstanceId, config.Id)
	if err != nil {
		log.Error("dbhelper", "getting datapoint by ID %v for config %v: %v", update.DatapointInstanceId, config.Id, err)
		return
	}

	// Alarm rule creation. This might be eventually moved to ontology sync.
	for i := range datapoint.Attributes {
		elionaAlarmID, err := eliona.CreateAlarm(datapoint.Asset.AssetID, datapoint.Subtype, datapoint.Attributes[i].Name, update.NeedAcknowledge, update.getPriority(), update.buildAlarmMessage())
		if err != nil {
			log.Error("eliona", "creating alarm: %v", err)
			return
		}
		if err := dbhelper.CreateAlarm(datapoint.Attributes[i].ID, elionaAlarmID, update.AlarmID); err != nil {
			log.Error("dbhelper", "creating alarm: %v", err)
			return
		}
	}

	alarms, err := dbhelper.GetAlarmsByOpenbosID(update.AlarmID)
	if err != nil {
		log.Error("dbhelper", "getting alarms for alarmID %s: %v", update.AlarmID, err)
		return
	}
	for _, alarm := range alarms {
		if err := eliona.UpdateAlarmStatus(alarm.ElionaAlarmID, update.Timestamp, update.Acked, update.getAckMessage(), update.Closed); err != nil {
			log.Error("eliona", "triggering alarm: %v", err)
			return
		}
	}
}

// ListenForOutputChanges listens to output attribute changes from Eliona.
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
			if err := outputData(output.AssetId, output.Data); err != nil {
				log.Error("dbhelper", "outputting data (%v) for assetId %v: %v", output.Data, output.AssetId, err)
				return
			}
		}
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
}

// outputData implements passing output data to broker.
func outputData(assetID int32, data map[string]interface{}) error {
	var attributesData []broker.AttributeData
	for name := range data {
		// Fetch the datapoint associated with the attribute name
		datapoint, err := dbhelper.GetDatapointByAttributeName(assetID, name)
		if err != nil {
			return fmt.Errorf("getting datapoint by assetID %v and name %v: %v", assetID, name, err)
		}

		// Fetch and format the latest data for all attributes of the datapoint
		latestData, err := formatComplexData(datapoint)
		if err != nil {
			return fmt.Errorf("formatting complex data for datapoint %v: %v", datapoint.ProviderID, err)
		}

		attributesData = append(attributesData, broker.AttributeData{
			Datapoint: datapoint,
			Value:     latestData,
		})
	}

	if len(attributesData) == 0 {
		return fmt.Errorf("shouldn't happen: no attribute data")
	}

	if err := broker.PutData(attributesData[0].Datapoint.Asset.Config, attributesData); err != nil {
		return fmt.Errorf("putting data: %v", err)
	}

	return nil
}

func formatComplexData(datapoint appmodel.Datapoint) (interface{}, error) {
	elionaAssetData, err := eliona.GetAssetData(datapoint.Asset.AssetID, datapoint.Subtype)
	if err != nil {
		return nil, fmt.Errorf("getting asset data: %v", err)
	}
	complexData := make(map[string]interface{})
	for _, attr := range datapoint.Attributes {
		value, ok := elionaAssetData.Data[attr.Name]
		if !ok {
			return nil, fmt.Errorf("data for '%s' not found in %+v", attr.Name, elionaAssetData.Data)
		}

		// Check if this is a nested attribute
		pathParts := strings.SplitN(attr.Name, ".", 2)
		if len(pathParts) > 1 {
			// Nested attribute: add to complex structure recursively
			if _, exists := complexData[pathParts[0]]; !exists {
				complexData[pathParts[0]] = make(map[string]interface{})
			}
			nestedData := complexData[pathParts[0]].(map[string]interface{})
			nestedData[pathParts[1]] = value
		} else {
			// Primitive attribute: directly set its value
			complexData[attr.Name] = value
		}
	}

	return complexData, nil
}

// ListenForAlarmChanges listens to output attribute changes from Eliona.
func ListenForAlarmChanges() {
	for { // We want to restart listening in case something breaks.
		outputs, err := eliona.ListenForAlarmChanges()
		if err != nil {
			log.Error("eliona", "listening for output changes: %v", err)
			return
		}
		for output := range outputs {
			if !output.AcknowledgeTimestamp.IsSet() {
				// We are only updating the acknowledges
				continue
			}
			alarm, err := dbhelper.GetAlarmByElionaID(output.RuleId)
			if errors.Is(err, dbhelper.ErrNotFound) {
				// Not from OpenBOS
				continue
			} else if err != nil {
				log.Error("dbhelper", "getting alarm by alarm ID: %v", err)
				return
			}
			config, err := dbhelper.GetConfigByElionaAlarmID(output.RuleId)
			if err != nil {
				log.Error("dbhelper", "getting config by alarm ID: %v", err)
				return
			}

			username, err := eliona.GetUserName(output.GetAcknowledgeUserId())
			if err != nil {
				log.Error("eliona", "getting ack username: %v", err)
				username = ""
			}
			if err := broker.AcknowledgeAlarm(config, alarm.OpenBOSAlarmID, username, output.GetAcknowledgeText()); err != nil {
				log.Error("broker", "acknowledging alarm: %v", err)
			}
		}
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
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

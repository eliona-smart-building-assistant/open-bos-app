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

package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/client"
	"github.com/jackc/pgx/v4"
	"net/http"
	apiserver "open-bos/api/generated"
	apiservices "open-bos/api/services"
	appmodel "open-bos/app/model"
	"open-bos/broker"
	"open-bos/complexdata"
	dbhelper "open-bos/db/helper"
	"open-bos/eliona"
	"strings"
	"sync"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v3"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/app"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

var appStatus = 0

const (
	statusOK = iota
	statusError
	statusFatal
)

func changeAppStatus(status int) {
	appStatus = status
	Heartbeat()
}

// Initialize TODO: MUST be replaced by new multi tenancy app concept
// Deprecated
func Initialize() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// TODO: Remove this quick fix later with new multi tenancy app concept (calling endpoint with HTTP encryption)
	initialized, err := getInitialized(ctx, conn, app.AppName())
	if err != nil {
		log.Fatal("init", "error during initialization: %w", err)
	}

	if !initialized {
		err := db.ExecFile(conn, "db/init.sql")
		if err != nil {
			log.Fatal("init", "error during initialization: %w", err)
		}

		_, err = conn.Exec(context.Background(), fmt.Sprintf("select fixprivilege('%s','%s')", strings.ReplaceAll(app.AppName(), "-", "_"), db.Username()))
		if err != nil {
			log.Fatal("init", "error during initialization: %w", err)
		}

		err = setInitialized(ctx, conn, app.AppName(), true)
		if err != nil {
			log.Fatal("init", "error during initialization: %w", err)
		}
	}
}

// getInitialized TODO: MUST be replaced by new multi tenancy app concept
// Deprecated
func getInitialized(ctx context.Context, conn *pgx.Conn, appName string) (bool, error) {
	var initialized bool
	err := conn.QueryRow(
		ctx,
		`SELECT initialized FROM public.eliona_store WHERE app_name = $1`,
		appName,
	).Scan(&initialized)

	if err != nil {
		return false, err
	}

	return initialized, nil
}

// setInitialized TODO: MUST be replaced by new multi tenancy app concept
// Deprecated
func setInitialized(ctx context.Context, conn *pgx.Conn, appName string, initialized bool) error {
	_, err := conn.Exec(
		ctx,
		`UPDATE public.eliona_store SET initialized = $1 WHERE app_name = $2`,
		initialized,
		appName,
	)

	return err
}

var notifyNoConfigsOnce sync.Once

func CollectData() {
	configs, err := dbhelper.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("dbhelper", "Couldn't read configs from DB: %v", err)
		changeAppStatus(statusFatal)
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
			_, err := dbhelper.SetConfigActiveState(context.Background(), config, true)
			if err != nil {
				return
			}
			log.Info("dbhelper", "Collecting initialized with Configuration %d:\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Site ID: %s\n",
				config.Id,
				config.Enable,
				config.RefreshInterval,
				config.RequestTimeout,
				config.ElionaSiteId)
		}

		common.RunOnceWithParam(func(config appmodel.Configuration) {
			log.Info("main", "Collecting %d started.", config.Id)
			if err := collectOntology(&config); err != nil {
				changeAppStatus(statusError)
				return // Error is handled in the method itself.
			}
			log.Info("main", "Collecting %d finished.", config.Id)

			log.Info("main", "Collecting alarms %d started.", config.Id)
			fetchAlarmRules(&config)
			log.Info("main", "Collecting alarms %d finished.", config.Id)

			if err := broker.SubscribeToOntologyChanges(config); err != nil {
				log.Error("broker", "subscribing to ontology changes: %v", err)
				changeAppStatus(statusError)
				return
			}
			log.Info("main", "Subscribed to ontology updates of config %d", config.Id)

			if err := broker.SubscribeToDataChanges(config); err != nil {
				log.Error("broker", "subscribing to data changes: %v", err)
				changeAppStatus(statusError)
				return
			}
			log.Info("main", "Subscribed to data updates of config %d", config.Id)

			if err := broker.SubscribeToAlarms(config); err != nil {
				log.Error("broker", "subscribing to alarm changes: %v", err)
				changeAppStatus(statusError)
				return
			}
			log.Info("main", "Subscribed to alarm updates of config %d", config.Id)
			changeAppStatus(statusOK)
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
	if err := collectOntology(&config); err != nil {
		return // Error is handled in the method itself.
	}
	log.Info("main", "Collecting %d finished.", config.Id)
}

func collectOntology(config *appmodel.Configuration) error {
	version, assetTypes, assets, err := broker.FetchOntology(*config)
	if errors.Is(err, broker.ErrNoUpdate) {
		log.Debug("broker", "ontology is up-to-date")
		return nil
	}
	if err != nil {
		log.Error("broker", "fetching assets: %v", err)
		return err
	}
	for _, assetType := range assetTypes {
		if err := asset.InitAssetType(client.ApiEndpointString(), config.ApiKey, assetType)(nil); err != nil {
			log.Error("eliona", "initializing asset type: %v", err)
			return err
		}
	}
	if err := eliona.CreateAssets(*config, assets); err != nil {
		log.Error("eliona", "creating assets: %v", err)
		return err
	}

	config.OntologyVersion = version
	err = dbhelper.UpdateConfigOntologyVersion(context.Background(), *config)
	if err != nil {
		return err
	}

	return nil
}

type AttributeDataUpdate struct {
	ConfigID            int64
	DatapointProviderID string
	Timestamp           time.Time
	Value               any
}

func UpdateDataPointInEliona(update AttributeDataUpdate) {
	config, err := dbhelper.GetConfig(context.Background(), update.ConfigID)
	if err != nil {
		log.Error("dbhelper", "Couldn't read config %d from DB: %v", update.ConfigID, err)
		return
	}
	if !config.Enable {
		if config.Active {
			_, _ = dbhelper.SetConfigActiveState(context.Background(), config, false)
		}
		return
	}
	if !config.Active {
		_, _ = dbhelper.SetConfigActiveState(context.Background(), config, true)
	}

	assetData := make(map[string]any)
	datapoint, err := dbhelper.GetDatapointById(update.DatapointProviderID, config.Id)
	if errors.Is(err, dbhelper.ErrNotFound) {
		log.Info("dbhelper", "datapoint not found (this may be caused by asset filter): %v", err)
		return
	}
	if err != nil {
		log.Error("dbhelper", "getting datapoint by ID %v for config %v: %v", update.DatapointProviderID, config.Id, err)
		return
	}
	// Complex decode support
	if complexData, ok := update.Value.(map[string]any); ok {
		decodedData := complexdata.DecodeComplexData(complexData, datapoint.AttributeNamePrefix)
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

	if err := eliona.UpsertAssetData(client.ApiEndpointString(), config.ApiKey, datapoint.Asset.AssetID, assetData, update.Timestamp, api.DataSubtype(datapoint.Subtype)); err != nil {
		log.Error("eliona", "upserting data: %v", err)
		return
	}
}

func fetchAlarmRules(config *appmodel.Configuration) {
	bosAlarms, err := broker.FetchAlarmRules(*config)
	if err != nil {
		log.Error("broker", "fetching alarm rules: %v", err)
		return
	}

	for _, bosAlarm := range bosAlarms {
		datapoint, err := dbhelper.GetDatapointById(bosAlarm.Datapoint.Identifier.DataPointInstanceID, config.Id)
		if errors.Is(err, dbhelper.ErrNotFound) {
			log.Info("dbhelper", "datapoint %v for alarm not found (this may be caused by asset filter): %v", bosAlarm.Datapoint.Identifier.DataPointInstanceID, err)
			continue
		}
		if err != nil {
			log.Error("dbhelper", "getting datapoint by ID %v for config %v: %v", bosAlarm.Datapoint.Identifier.DataPointInstanceID, config.Id, err)
			return
		}
		prio, err := translateSeverity(bosAlarm.Template.Trigger.Severity)
		if err != nil {
			log.Error("broker", "translating severity: %v", err)
			prio = api.ALARM_PRIORITY_HEIGHT
		}
		presentAlarmRules, err := dbhelper.GetAlarmsByOpenbosID(bosAlarm.ID)
		if err != nil {
			log.Error("dbhelper", "getting alarms for alarmID %s: %v", bosAlarm.ID, err)
			return
		}
		for _, attribute := range datapoint.Attributes {
			ruleExists := false
			for _, presentAlarmRule := range presentAlarmRules {
				if presentAlarmRule.OpenBOSAlarmID == bosAlarm.ID && presentAlarmRule.ElionaAttributeID == attribute.ID {
					// Rule already exists. Just update it.
					elionaAlarmID, err := eliona.UpdateAlarm(config.ApiKey, presentAlarmRule.ElionaAlarmID, datapoint.Asset.AssetID, datapoint.Subtype, attribute.Name, prio, bosAlarm.Template.Name, buildAlarmMessage(bosAlarm))
					if err != nil {
						log.Error("eliona", "creating alarm: %v", err)
						return
					}
					log.Debug("app", "updated alarm %v", elionaAlarmID)
					ruleExists = true
					break
				}
			}
			if ruleExists {
				continue
			}
			elionaAlarmID, err := eliona.CreateAlarm(config.ApiKey, datapoint.Asset.AssetID, datapoint.Subtype, attribute.Name, prio, bosAlarm.Template.Name, buildAlarmMessage(bosAlarm))
			if err != nil {
				log.Error("eliona", "creating alarm: %v", err)
				return
			}
			if err := dbhelper.CreateAlarm(attribute.ID, elionaAlarmID, bosAlarm.ID); err != nil {
				log.Error("dbhelper", "creating alarm: %v", err)
				return
			}
			log.Debug("app", "created alarm %v", elionaAlarmID)
		}
	}
}

func translateSeverity(bosSeverity string) (api.AlarmPriority, error) {
	var priority api.AlarmPriority
	switch bosSeverity {
	case "Log":
		priority = api.ALARM_PRIORITY_INFO
	case "Low":
		priority = api.ALARM_PRIORITY_LOW
	case "High":
		priority = api.ALARM_PRIORITY_HEIGHT // todo: Or should it be "medium"?
	case "Urgent", "Critical":
		priority = api.ALARM_PRIORITY_HEIGHT
	default:
		return 0, fmt.Errorf("unknown severity '%s'", bosSeverity)
	}
	return priority, nil
}

// buildAlarmMessage builds alarm message to Eliona format.
func buildAlarmMessage(alarm broker.AlarmRule) map[string]interface{} {
	message := make(map[string]interface{})

	languageCodes := []string{"de", "en", "fr", "it"}

	var descriptionPart string
	if alarm.Template.Trigger.Description != "" {
		descriptionPart = fmt.Sprintf(": %s", alarm.Template.Trigger.Description)
	}
	templateCome := fmt.Sprintf("%s%s {{asset.name}} ({{alarm.val}})", alarm.Template.Name, descriptionPart)

	come := make(map[string]string)
	for _, lang := range languageCodes {
		come[lang] = templateCome
	}

	goneTranslations := map[string]string{
		"de": fmt.Sprintf("%s behoben", alarm.Template.Name),
		"en": fmt.Sprintf("%s resolved", alarm.Template.Name),
		"fr": fmt.Sprintf("%s résolu", alarm.Template.Name),
		"it": fmt.Sprintf("%s risolto", alarm.Template.Name),
	}

	gone := make(map[string]string)
	for _, lang := range languageCodes {
		gone[lang] = goneTranslations[lang]
	}

	message["come"] = come
	message["gone"] = gone

	return message
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

	alarms, err := dbhelper.GetAlarmsByOpenbosID(update.AlarmID)
	if err != nil {
		log.Error("dbhelper", "getting alarms for alarmID %s: %v", update.AlarmID, err)
		return
	}
	message := buildAlarmUpdateMessage(update)
	for _, alarm := range alarms {
		if err := eliona.UpdateAlarmStatus(config.ApiKey, alarm.ElionaAlarmID, message, update.Timestamp, update.Acked, update.getAckMessage(), update.Closed); err != nil {
			log.Error("eliona", "triggering alarm: %v", err)
			return
		}
	}
}

func buildAlarmUpdateMessage(alarm AlarmUpdate) map[string]interface{} {
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

func ListenForOutputChanges() {
	for _, apiKey := range dbhelper.FetchedApiKeys {
		go listenForOutputChanges(apiKey)
	}
}

// listenForOutputChanges listens to output attribute changes from Eliona.
func listenForOutputChanges(apiKey string) {
	for { // We want to restart listening in case something breaks.
		outputs, err := eliona.ListenForOutputChanges(apiKey)
		if err != nil {
			log.Error("eliona", "listening for output changes with key %s: %v", apiKey, err)
			time.Sleep(time.Second * 5) // avoid tight error loop
			continue
		}
		for output := range outputs {
			if cr := output.ClientReference.Get(); cr != nil && *cr == eliona.ClientReference {
				// Just an echoed value this app sent.
				continue
			}
			if err := outputData(apiKey, output.AssetId, output.Data); err != nil {
				log.Error("dbhelper", "outputting data (%v) for assetId %v with key %s: %v",
					output.Data, output.AssetId, apiKey, err)
				continue
			}
		}
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
}

// outputData implements passing output data to broker.
func outputData(apiKey string, assetID int32, data map[string]interface{}) error {
	var attributesData []broker.AttributeData
	for name, value := range data {
		// Fetch the datapoint associated with the attribute name
		datapoint, err := dbhelper.GetDatapointByAttributeName(assetID, name)
		if errors.Is(err, dbhelper.ErrNotFound) {
			// Asset not originating from this app.
			continue
		}
		if err != nil {
			return fmt.Errorf("getting datapoint by assetID %v and name %v: %v", assetID, name, err)
		}

		var latestData any
		if len(datapoint.Attributes) == 1 {
			latestData = value
		} else {
			// Fetch and format the latest data for all attributes of the datapoint
			latestData, err = formatComplexData(apiKey, datapoint)
			if err != nil {
				return fmt.Errorf("formatting complex data for datapoint %v: %v", datapoint.ProviderID, err)
			}
		}

		attributesData = append(attributesData, broker.AttributeData{
			Datapoint: datapoint,
			Value:     latestData,
		})
	}

	if len(attributesData) == 0 {
		return nil
	}

	if err := broker.PutData(attributesData[0].Datapoint.Asset.Config, attributesData); err != nil {
		return fmt.Errorf("putting data: %v", err)
	}

	return nil
}

func formatComplexData(apiKey string, datapoint appmodel.Datapoint) (interface{}, error) {
	elionaAssetData, err := eliona.GetAssetData(client.ApiEndpointString(), apiKey, datapoint.Asset.AssetID, datapoint.Subtype)
	if err != nil {
		return nil, fmt.Errorf("getting asset data: %v", err)
	}

	complexData := make(map[string]interface{})
	for _, attr := range datapoint.Attributes {
		value, ok := elionaAssetData.Data[attr.Name]
		if !ok {
			return nil, fmt.Errorf("data for '%s' not found in %+v", attr.Name, elionaAssetData.Data)
		}

		pathParts := strings.SplitN(attr.Name, ".", 2)
		// Check if this is a nested attribute
		if len(pathParts) < 2 {
			return nil, fmt.Errorf("inconsistency: not a nested attribute")
		}
		complexData[pathParts[1]] = value
	}

	return complexData, nil
}

func ListenForAlarmChanges() {
	for _, apiKey := range dbhelper.FetchedApiKeys {
		go listenForAlarmChanges(apiKey)
	}
}

// listenForAlarmChanges listens to output attribute changes from Eliona.
func listenForAlarmChanges(apiKey string) {
	for { // We want to restart listening in case something breaks.
		apiAlarms, err := eliona.ListenForAlarmChanges(apiKey)
		if err != nil {
			log.Error("eliona", "listening for alarm changes: %v", err)
			time.Sleep(time.Second * 5) // avoid tight error loop
			continue
		}
		for apiAlarm := range apiAlarms {
			if !apiAlarm.AcknowledgeTimestamp.IsSet() {
				// We are only updating the acknowledges
				continue
			}
			alarm, err := dbhelper.GetAlarmByElionaID(apiAlarm.RuleId)
			if errors.Is(err, dbhelper.ErrNotFound) {
				// Not from OpenBOS
				continue
			} else if err != nil {
				log.Error("dbhelper", "getting alarm by alarm ID: %v", err)
				return
			}
			config, err := dbhelper.GetConfigByElionaAlarmID(apiAlarm.RuleId)
			if err != nil {
				log.Error("dbhelper", "getting config by alarm ID: %v", err)
				return
			}

			username, err := eliona.GetUserName(apiKey, apiAlarm.GetAcknowledgeUserId())
			if err != nil {
				log.Error("eliona", "getting ack username: %v", err)
				username = ""
			}
			if err := broker.AcknowledgeAlarm(config, alarm.OpenBOSAlarmID, username, apiAlarm.GetAcknowledgeText()); err != nil {
				log.Error("broker", "acknowledging alarm: %v", err)
			}
		}
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
}

func Heartbeat() {
	for _, apiKey := range dbhelper.FetchedApiKeys {
		go heartbeat(apiKey)
	}
}

func heartbeat(apiKey string) {
	root, err := dbhelper.GetRootAsset()
	if errors.Is(err, dbhelper.ErrNotFound) {
		// No root yet, nothing to do
		return
	}
	if err != nil {
		log.Error("dbhelper", "getting root assets: %v", err)
		return
	}

	if err := eliona.UpsertData(client.ApiEndpointString(), apiKey, root.AssetID, map[string]any{"status": appStatus}, time.Now(), api.STATUS); err != nil {
		log.Error("eliona", "upserting data as heartbeat: %v", err)
		return
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
	changeAppStatus(statusFatal)
}

func Teardown() {
	configs, err := dbhelper.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("dbhelper", "Couldn't read configs from DB: %v", err)
		changeAppStatus(statusFatal)
		return
	}

	for _, config := range configs {
		if err := broker.CancelSubscriptions(config); err != nil {
			log.Error("broker", "cancelling all subscriptions: %v", err)
			return
		}
	}
}

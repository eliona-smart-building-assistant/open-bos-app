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

package dbhelper

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	"log"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	appmodel "open-bos/v2/app/model"

	dbgen "open-bos/v2/db/generated"

	"github.com/eliona-smart-building-assistant/go-eliona/v2/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

func InsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}
	if err := dbConfig.InsertG(ctx, boil.Infer()); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}
	result, err := toAppConfig(&dbConfig)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("converting back to appConfig: %v", err)
	}
	return result, nil
}

func UpsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}
	if err := dbConfig.UpsertG(ctx, true, []string{"id"}, boil.Blacklist("id"), boil.Infer()); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}
	result, err := toAppConfig(&dbConfig)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("converting back to appConfig: %v", err)
	}
	return result, nil
}

func UpdateConfigOntologyVersion(ctx context.Context, config appmodel.Configuration) error {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("creating DB config from App config: %v", err)
	}
	if _, err := dbConfig.UpdateG(ctx, boil.Whitelist(dbgen.ConfigurationColumns.OntologyVersion)); err != nil {
		return fmt.Errorf("updating DB config ontology version: %v", err)
	}
	return nil
}

func GetConfig(ctx context.Context, configID int64) (appmodel.Configuration, error) {
	dbConfig, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(configID),
	).OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Configuration{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("fetching config from database: %v", err)
	}
	appConfig, err := toAppConfig(dbConfig)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating App config from DB config: %v", err)
	}
	return appConfig, nil
}

func DeleteConfig(ctx context.Context, configID int64) error {
	if _, err := dbgen.Assets(
		dbgen.AssetWhere.ConfigurationID.EQ(configID),
	).DeleteAllG(ctx); err != nil {
		return fmt.Errorf("deleting assets from database: %v", err)
	}
	count, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(configID),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("deleting config from database: %v", err)
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) configs by ID", count)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func toDbConfig(ctx context.Context, appConfig appmodel.Configuration) (dbConfig dbgen.Configuration, err error) {
	dbConfig.ElionaTenantID = appConfig.ElionaTenantId
	dbConfig.ElionaSiteID = appConfig.ElionaSiteId
	dbConfig.GWID = appConfig.GwId
	dbConfig.ClientID = appConfig.ClientId
	dbConfig.ClientSecret = appConfig.ClientSecret
	dbConfig.OntologyVersion = appConfig.OntologyVersion
	dbConfig.AppPublicAPIURL = appConfig.AppPublicApiUrl

	dbConfig.ID = appConfig.Id
	dbConfig.RefreshInterval = appConfig.RefreshInterval
	dbConfig.RequestTimeout = appConfig.RequestTimeout
	af, err := json.Marshal(appConfig.AssetFilter)
	if err != nil {
		return dbgen.Configuration{}, fmt.Errorf("marshalling assetFilter: %v", err)
	}
	dbConfig.AssetFilter = af
	dbConfig.Active = appConfig.Active
	dbConfig.Enable = appConfig.Enable
	dbConfig.UserID = null.StringFromPtr(appConfig.UserId)

	if env := frontend.GetEnvironment(ctx); env != nil {
		dbConfig.UserID = null.StringFrom(env.UserId)

		if appConfig.AppPublicApiUrl == "" {
			dbConfig.AppPublicAPIURL = fmt.Sprintf("%s/apps-public/open-bos", env.Iss)
		}
	}

	return dbConfig, nil
}

func toAppConfig(dbConfig *dbgen.Configuration) (appConfig appmodel.Configuration, err error) {
	appConfig.ElionaTenantId = dbConfig.ElionaTenantID
	appConfig.ElionaSiteId = dbConfig.ElionaSiteID
	appConfig.GwId = dbConfig.GWID
	appConfig.ClientId = dbConfig.ClientID
	appConfig.ClientSecret = dbConfig.ClientSecret
	appConfig.OntologyVersion = dbConfig.OntologyVersion
	appConfig.AppPublicApiUrl = dbConfig.AppPublicAPIURL

	appConfig.Id = dbConfig.ID
	appConfig.Enable = dbConfig.Enable
	appConfig.RefreshInterval = dbConfig.RefreshInterval
	appConfig.RequestTimeout = dbConfig.RequestTimeout
	var af [][]appmodel.FilterRule
	if err := json.Unmarshal(dbConfig.AssetFilter, &af); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("unmarshalling assetFilter: %v", err)
	}
	appConfig.AssetFilter = af
	appConfig.Active = dbConfig.Active
	appConfig.UserId = dbConfig.UserID.Ptr()

	// TODO: MUST be replaced by new multi tenancy app concept
	if apiKey, ok := FetchedApiKeys[dbConfig.ElionaTenantID]; ok {
		appConfig.ApiKey = apiKey
	} else {
		log.Fatal("conf", "api key not found in DB for: ", dbConfig.ElionaTenantID)
	}

	return appConfig, nil
}

// FetchedApiKeys TODO: MUST be replaced by new multi tenancy app concept
// Deprecated
var FetchedApiKeys map[string]string

// FetchApiKeys TODO: MUST be replaced by new multi tenancy app concept
// Deprecated
func FetchApiKeys(appName string) {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, appName)
	defer conn.Close(ctx)

	apiKeys := make(map[string]string)

	rows, err := conn.Query(
		ctx,
		`SELECT tenant_id, api_key
		 FROM public.eliona_app
		 JOIN open_bos.configuration on (tenant_id = eliona_tenant_id)
		 WHERE app_name = $1`,
		appName,
	)
	if err != nil {
		log.Fatal("init", "error during fetching api keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tenantID, apiKey string
		if err := rows.Scan(&tenantID, &apiKey); err != nil {
			log.Fatal("init", "error during fetching api keys: %w", err)
		}
		apiKeys[tenantID] = apiKey
	}

	if rows.Err() != nil {
		log.Fatal("init", "error during fetching api keys: %w", rows.Err())
	}

	// If no rows, apiKeys will just be empty map
	FetchedApiKeys = apiKeys
}

func GetConfigs(ctx context.Context) ([]appmodel.Configuration, error) {
	dbConfigs, err := dbgen.Configurations().AllG(ctx)
	if err != nil {
		return nil, err
	}
	var appConfigs []appmodel.Configuration
	for _, dbConfig := range dbConfigs {
		ac, err := toAppConfig(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("creating App config from DB config: %v", err)
		}
		appConfigs = append(appConfigs, ac)
	}
	return appConfigs, nil
}

func SetConfigActiveState(ctx context.Context, config appmodel.Configuration, state bool) (int64, error) {
	return dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(config.Id),
	).UpdateAllG(ctx, dbgen.M{
		dbgen.ConfigurationColumns.Active: state,
	})
}

func SetAllConfigsInactive(ctx context.Context) (int64, error) {
	return dbgen.Configurations().UpdateAllG(ctx, dbgen.M{
		dbgen.ConfigurationColumns.Active: false,
	})
}

func InsertAsset(ctx context.Context, config appmodel.Configuration, globalAssetID string, assetId int32, providerId string, isRootSpace bool) (assetID int64, err error) {
	var dbAsset dbgen.Asset
	dbAsset.ConfigurationID = config.Id
	dbAsset.GlobalAssetID = globalAssetID
	dbAsset.AssetID = null.Int32From(assetId)
	dbAsset.ProviderID = providerId
	dbAsset.IsRootSpace = isRootSpace
	if err := dbAsset.InsertG(ctx, boil.Infer()); err != nil {
		return 0, fmt.Errorf("inserting asset: %v", err)
	}

	return dbAsset.ID, nil
}

func GetAssetId(ctx context.Context, config appmodel.Configuration, globalAssetID string) (*int32, error) {
	dbAsset, err := dbgen.Assets(
		dbgen.AssetWhere.ConfigurationID.EQ(config.Id),
		dbgen.AssetWhere.GlobalAssetID.EQ(globalAssetID),
	).AllG(ctx)
	if err != nil || len(dbAsset) == 0 {
		return nil, err
	}
	return common.Ptr(dbAsset[0].AssetID.Int32), nil
}

func InsertAssetAttributes(ctx context.Context, assetId int64, datapoints []appmodel.Datapoint) error {
	for _, datapoint := range datapoints {
		// Insert OpenBOS Datapoint
		dbDatapoint := dbgen.OpenbosDatapoint{
			AssetID:    assetId,
			Subtype:    datapoint.Subtype,
			ProviderID: datapoint.ProviderID,
			Name:       datapoint.AttributeNamePrefix,
		}

		var existing dbgen.OpenbosDatapoint

		err := dbgen.OpenbosDatapoints(
			dbgen.OpenbosDatapointWhere.ProviderID.EQ(datapoint.ProviderID),
		).BindG(ctx, &existing)

		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("checking existing datapoint: %v", err)
		}

		// If a record exists, compare other fields
		if existing.ID != 0 {
			if existing.Subtype == datapoint.Subtype && existing.Name == datapoint.AttributeNamePrefix {
				// Fields are identical, so ignore insertion
				continue
			} else {
				// Fields are different, return an error
				return fmt.Errorf("conflict: datapoint with provider_id %s exists but has different values: %+v", datapoint.ProviderID, existing)
			}
		}

		// Insert if no existing record
		if err := dbDatapoint.InsertG(ctx, boil.Infer()); err != nil {
			return fmt.Errorf("inserting datapoint %+v: %v", datapoint, err)
		}

		// Insert associated Eliona Attributes for the Datapoint
		for _, attribute := range datapoint.Attributes {
			dbAttribute := dbgen.ElionaAttribute{
				OpenbosDatapointID:  dbDatapoint.ID,
				ElionaAttributeName: attribute.Name,
			}

			if err := dbAttribute.InsertG(ctx, boil.Infer()); err != nil {
				return fmt.Errorf("inserting attribute %+v for datapoint %v: %v", attribute, datapoint.ProviderID, err)
			}
		}
	}

	return nil
}

func toAppAsset(dbAsset dbgen.Asset, config appmodel.Configuration) appmodel.Asset {
	return appmodel.Asset{
		ID:            dbAsset.ID,
		Config:        config,
		GlobalAssetID: dbAsset.GlobalAssetID,
		ProviderID:    dbAsset.ProviderID,
		AssetID:       dbAsset.AssetID.Int32,
	}
}

func GetAssetById(assetId int32) (appmodel.Asset, error) {
	ctx := context.Background()
	asset, err := dbgen.Assets(
		dbgen.AssetWhere.AssetID.EQ(null.Int32From(assetId)),
	).OneG(ctx)
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}
	if !asset.AssetID.Valid {
		return appmodel.Asset{}, fmt.Errorf("shouldn't happen: assetID is nil")
	}
	c, err := asset.Configuration().OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Asset{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching configuration: %v", err)
	}
	config, err := toAppConfig(c)
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("translating configuration: %v", err)
	}
	return toAppAsset(*asset, config), nil
}

func GetRootAsset() (appmodel.Asset, error) {
	asset, err := dbgen.Assets(
		dbgen.AssetWhere.ProviderID.EQ("root"),
	).OneG(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Asset{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching root assets: %v", err)
	}

	c, err := asset.Configuration().OneG(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Asset{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching configuration: %v", err)
	}
	config, err := toAppConfig(c)
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("translating configuration: %v", err)
	}
	return toAppAsset(*asset, config), nil
}

func GetDatapointById(providerDatapointID string, configID int64) (appmodel.Datapoint, error) {
	ctx := context.Background()

	datapointTable := "open_bos.openbos_datapoint"
	assetTable := "open_bos.asset"
	configTable := "open_bos.configuration"

	// Query to fetch the OpenBOS datapoint and its associated asset
	datapoint, err := dbgen.OpenbosDatapoints(
		qm.InnerJoin(fmt.Sprintf("%s ON %s.asset_id = %s.id", assetTable, datapointTable, assetTable)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s.id = %s.configuration_id", configTable, configTable, assetTable)),
		dbgen.ConfigurationWhere.ID.EQ(configID),
		dbgen.OpenbosDatapointWhere.ProviderID.EQ(providerDatapointID),
	).OneG(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appmodel.Datapoint{}, fmt.Errorf("no datapoint found for provider ID %v in config %v: %w", providerDatapointID, configID, ErrNotFound)
		}
		return appmodel.Datapoint{}, fmt.Errorf("fetching datapoint for provider ID %v: %v", providerDatapointID, err)
	}

	// Fetch associated attributes for the datapoint
	attributes, err := dbgen.ElionaAttributes(
		dbgen.ElionaAttributeWhere.OpenbosDatapointID.EQ(datapoint.ID),
	).AllG(ctx)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching attributes for datapoint ID %v: %v", datapoint.ID, err)
	}

	// Map attributes to the appmodel structure
	var appAttributes []appmodel.Attribute
	for _, attr := range attributes {
		appAttributes = append(appAttributes, appmodel.Attribute{
			ID:   attr.ID,
			Name: attr.ElionaAttributeName,
		})
	}

	// Fetch the associated asset
	asset, err := datapoint.Asset().OneG(ctx)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching asset for datapoint ID %v: %v", datapoint.ID, err)
	}

	c, err := asset.Configuration().OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Datapoint{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching configuration: %v", err)
	}
	config, err := toAppConfig(c)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("translating configuration: %v", err)
	}
	appAsset := toAppAsset(*asset, config)

	return appmodel.Datapoint{
		ProviderID:          datapoint.ProviderID,
		Subtype:             datapoint.Subtype,
		Asset:               &appAsset,
		AttributeNamePrefix: datapoint.Name,
		Attributes:          appAttributes,
	}, nil
}

func GetDatapointByAttributeName(assetID int32, attributeName string) (appmodel.Datapoint, error) {
	ctx := context.Background()

	// Define table names for readability
	attributeTable := "open_bos.eliona_attribute"
	datapointTable := "open_bos.openbos_datapoint"
	assetTable := "open_bos.asset"
	configTable := "open_bos.configuration"

	// Query to find the attribute and join with the datapoint to which it belongs
	attribute, err := dbgen.ElionaAttributes(
		qm.InnerJoin(fmt.Sprintf("%s ON %s.id = %s.openbos_datapoint_id", datapointTable, datapointTable, attributeTable)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s.asset_id = %s.id", assetTable, datapointTable, assetTable)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s.id = %s.configuration_id", configTable, configTable, assetTable)),
		dbgen.AssetWhere.AssetID.EQ(null.Int32From(assetID)),
		dbgen.ElionaAttributeWhere.ElionaAttributeName.EQ(attributeName),
	).OneG(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appmodel.Datapoint{}, fmt.Errorf("no attribute found for asset ID %v and attribute name %v: %w", assetID, attributeName, ErrNotFound)
		}
		return appmodel.Datapoint{}, fmt.Errorf("fetching attribute: %v", err)
	}

	// Retrieve the associated datapoint
	datapoint, err := attribute.OpenbosDatapoint().OneG(ctx)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching datapoint for attribute ID %v: %v", attribute.ID, err)
	}

	// Retrieve all attributes related to the datapoint
	relatedAttributes, err := dbgen.ElionaAttributes(
		dbgen.ElionaAttributeWhere.OpenbosDatapointID.EQ(datapoint.ID),
	).AllG(ctx)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching related attributes for datapoint ID %v: %v", datapoint.ID, err)
	}

	// Map attributes to appmodel.Attribute
	var appAttributes []appmodel.Attribute
	for _, attr := range relatedAttributes {
		appAttributes = append(appAttributes, appmodel.Attribute{
			ID:   attr.ID,
			Name: attr.ElionaAttributeName,
		})
	}

	// Retrieve the associated asset
	asset, err := datapoint.Asset().OneG(ctx)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching asset for datapoint ID %v: %v", datapoint.ID, err)
	}

	// Retrieve the configuration associated with the asset
	config, err := asset.Configuration().OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Datapoint{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("fetching configuration for asset ID %v: %v", asset.ID, err)
	}
	appConfig, err := toAppConfig(config)
	if err != nil {
		return appmodel.Datapoint{}, fmt.Errorf("translating configuration: %v", err)
	}
	appAsset := toAppAsset(*asset, appConfig)

	// Construct and return the Datapoint object
	return appmodel.Datapoint{
		ProviderID:          datapoint.ProviderID,
		Subtype:             datapoint.Subtype,
		Asset:               &appAsset,
		AttributeNamePrefix: datapoint.Name,
		Attributes:          appAttributes,
	}, nil
}

func CreateAlarm(attributeID int64, elionaAlarmID int32, openbosAlarmID string) error {
	dbAlarm := dbgen.Alarm{
		ElionaAttributeID: attributeID,
		ElionaAlarmID:     elionaAlarmID,
		OpenbosAlarmID:    openbosAlarmID,
	}
	return dbAlarm.UpsertG(context.Background(), true, []string{dbgen.AlarmColumns.ElionaAlarmID}, boil.Infer(), boil.Infer())
}

func GetAlarmsByOpenbosID(openbosID string) ([]appmodel.Alarm, error) {
	alarms, err := dbgen.Alarms(
		dbgen.AlarmWhere.OpenbosAlarmID.EQ(openbosID),
	).AllG(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching alarms: %v", err)
	}
	var appAlarms []appmodel.Alarm
	for _, alarm := range alarms {
		appAlarms = append(appAlarms, appmodel.Alarm{
			ElionaAlarmID:     alarm.ElionaAlarmID,
			ElionaAttributeID: alarm.ElionaAttributeID,
			OpenBOSAlarmID:    alarm.OpenbosAlarmID,
		})
	}
	return appAlarms, nil
}

func GetConfigByElionaAlarmID(elionaAlarmID int32) (appmodel.Configuration, error) {
	ctx := context.Background()

	alarmTable := "open_bos.alarm"
	attributeTable := "open_bos.eliona_attribute"
	datapointTable := "open_bos.openbos_datapoint"
	assetTable := "open_bos.asset"
	configTable := "open_bos.configuration"

	config, err := dbgen.Configurations(
		qm.InnerJoin(fmt.Sprintf("%s ON %s.id = %s.configuration_id", assetTable, configTable, assetTable)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s.asset_id = %s.id", datapointTable, datapointTable, assetTable)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s.id = %s.openbos_datapoint_id", attributeTable, datapointTable, attributeTable)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s.id = %s.eliona_attribute_id", alarmTable, attributeTable, alarmTable)),
		dbgen.AlarmWhere.ElionaAlarmID.EQ(elionaAlarmID),
	).OneG(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appmodel.Configuration{}, ErrNotFound
		}
		return appmodel.Configuration{}, fmt.Errorf("fetching attribute: %v", err)
	}

	appConfig, err := toAppConfig(config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("translating configuration: %v", err)
	}

	return appConfig, nil
}

func GetAlarmByElionaID(elionaAlarmID int32) (appmodel.Alarm, error) {
	ctx := context.Background()

	dbAlarm, err := dbgen.Alarms(
		dbgen.AlarmWhere.ElionaAlarmID.EQ(elionaAlarmID),
	).OneG(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appmodel.Alarm{}, ErrNotFound
		}
		return appmodel.Alarm{}, fmt.Errorf("fetching attribute: %v", err)
	}

	return appmodel.Alarm{
		ElionaAlarmID:  elionaAlarmID,
		OpenBOSAlarmID: dbAlarm.OpenbosAlarmID,
	}, nil
}

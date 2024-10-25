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

package dbhelper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	appmodel "open-bos/app/model"

	dbgen "open-bos/db/generated"

	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
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
	return config, nil
}

func UpsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}
	if err := dbConfig.UpsertG(ctx, true, []string{"id"}, boil.Blacklist("id"), boil.Infer()); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}
	return config, nil
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
	dbConfig.Gwid = appConfig.Gwid
	dbConfig.ClientID = appConfig.ClientID
	dbConfig.ClientSecret = appConfig.ClientSecret
	dbConfig.OntologyVersion = appConfig.OntologyVersion
	dbConfig.AppPublicAPIURL = appConfig.AppPublicAPIURL

	dbConfig.ID = appConfig.Id
	dbConfig.RefreshInterval = appConfig.RefreshInterval
	dbConfig.RequestTimeout = appConfig.RequestTimeout
	dbConfig.Active = appConfig.Active
	dbConfig.Enable = appConfig.Enable
	dbConfig.ProjectIds = appConfig.ProjectIDs

	if env := frontend.GetEnvironment(ctx); env != nil {
		dbConfig.UserID = env.UserId

		if appConfig.AppPublicAPIURL == "" {
			dbConfig.AppPublicAPIURL = fmt.Sprintf("%s/apps-public/open-bos", env.Iss)
		}
	}

	return dbConfig, nil
}

func toAppConfig(dbConfig *dbgen.Configuration) (appConfig appmodel.Configuration, err error) {
	appConfig.Gwid = dbConfig.Gwid
	appConfig.ClientID = dbConfig.ClientID
	appConfig.ClientSecret = dbConfig.ClientSecret
	appConfig.OntologyVersion = dbConfig.OntologyVersion
	appConfig.AppPublicAPIURL = dbConfig.AppPublicAPIURL

	appConfig.Id = dbConfig.ID
	appConfig.Enable = dbConfig.Enable
	appConfig.RefreshInterval = dbConfig.RefreshInterval
	appConfig.RequestTimeout = dbConfig.RequestTimeout
	appConfig.Active = dbConfig.Active
	appConfig.ProjectIDs = dbConfig.ProjectIds
	appConfig.UserId = dbConfig.UserID
	return appConfig, nil
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

func InsertAsset(ctx context.Context, config appmodel.Configuration, projId string, globalAssetID string, assetId int32, providerId string) (assetID int64, err error) {
	var dbAsset dbgen.Asset
	dbAsset.ConfigurationID = config.Id
	dbAsset.ProjectID = projId
	dbAsset.GlobalAssetID = globalAssetID
	dbAsset.AssetID = null.Int32From(assetId)
	dbAsset.ProviderID = providerId
	if err := dbAsset.InsertG(ctx, boil.Infer()); err != nil {
		return 0, fmt.Errorf("inserting asset: %v", err)
	}

	// TODO: Is this needed?
	// if err := dbAsset.ReloadG(ctx); err != nil {
	// 	return 0, fmt.Errorf("reloading asset: %v", err)
	// }
	return dbAsset.ID, nil
}

func GetAssetId(ctx context.Context, config appmodel.Configuration, projId string, globalAssetID string) (*int32, error) {
	dbAsset, err := dbgen.Assets(
		dbgen.AssetWhere.ConfigurationID.EQ(config.Id),
		dbgen.AssetWhere.ProjectID.EQ(projId),
		dbgen.AssetWhere.GlobalAssetID.EQ(globalAssetID),
	).AllG(ctx)
	if err != nil || len(dbAsset) == 0 {
		return nil, err
	}
	return common.Ptr(dbAsset[0].AssetID.Int32), nil
}

func InsertAssetAttributes(ctx context.Context, assetId int64, attributes []appmodel.Attribute) error {
	for _, attribute := range attributes {
		dbAttribute := dbgen.Attribute{
			AssetID:    assetId,
			Subtype:    attribute.Subtype,
			Name:       attribute.Name,
			ProviderID: attribute.ProviderID,
		}
		if err := dbAttribute.InsertG(ctx, boil.Infer()); err != nil {
			return fmt.Errorf("inserting attribute %+v: %v", attribute, err)
		}
	}
	return nil
}

func toAppAsset(ctx context.Context, dbAsset dbgen.Asset, config appmodel.Configuration) (appmodel.Asset, error) {
	// dbSubtypes, err := dbAsset.AssetSubtypes().AllG(ctx)
	// if err != nil {
	// 	return appmodel.Asset{}, fmt.Errorf("fetching asset subtypes: %v", err)
	// }
	// var subtypes []appmodel.AssetSubtype
	// for _, dbs := range dbSubtypes {
	// 	var data map[string]any
	// 	if err := dbs.Data.Unmarshal(&data); err != nil {
	// 		return appmodel.Asset{}, fmt.Errorf("unmarshalling: %v \nData: %s", err, dbs.Data)
	// 	}
	// 	subtypes = append(subtypes, appmodel.AssetSubtype{
	// 		ID:      dbs.ID,
	// 		Subtype: dbs.Subtype,
	// 		Data:    data,
	// 	})
	// }
	return appmodel.Asset{
		ID:            dbAsset.ID,
		Config:        config,
		ProjectID:     dbAsset.ProjectID,
		GlobalAssetID: dbAsset.GlobalAssetID,
		ProviderID:    dbAsset.ProviderID,
		AssetID:       dbAsset.AssetID.Int32,
		//Subtypes:      subtypes,
	}, nil
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
	appAsset, err := toAppAsset(ctx, *asset, config)
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("converting to app asset: %v", err)
	}
	return appAsset, nil
}

func GetAttributeById(providerID string, configID int64) (appmodel.Attribute, error) {
	ctx := context.Background()
	attribute, err := dbgen.Attributes(
		qm.InnerJoin(dbgen.TableNames.Asset+" a on "+dbgen.TableNames.Attribute+"."+dbgen.AttributeColumns.AssetID+"=a.id"),
		qm.InnerJoin(dbgen.TableNames.Configuration+" c on "+"a."+dbgen.AssetColumns.ConfigurationID+"=c.id"),
		dbgen.ConfigurationWhere.ID.EQ(configID),
		dbgen.AttributeWhere.ProviderID.EQ(providerID),
	).OneG(ctx)
	if err != nil {
		return appmodel.Attribute{}, fmt.Errorf("fetching attribute: %v", err)
	}

	a, err := attribute.Asset().OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Attribute{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Attribute{}, fmt.Errorf("fetching asset: %v", err)
	}

	c, err := a.Configuration().OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Attribute{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Attribute{}, fmt.Errorf("fetching configuration: %v", err)
	}
	config, err := toAppConfig(c)
	if err != nil {
		return appmodel.Attribute{}, fmt.Errorf("translating configuration: %v", err)
	}
	asset, err := toAppAsset(ctx, *a, config)
	if err != nil {
		return appmodel.Attribute{}, fmt.Errorf("converting to app asset: %v", err)
	}

	appAttribute := appmodel.Attribute{
		Name:       attribute.Name,
		Subtype:    attribute.Subtype,
		ProviderID: attribute.ProviderID,
		Asset:      &asset,
	}
	return appAttribute, nil
}

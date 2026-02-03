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
	appmodel "weather-app2/v2/app/model"
	dbgen "weather-app2/v2/db/generated"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/app"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

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

func GetConfig(ctx context.Context, configID int64) (appmodel.Configuration, error) {
	dbConfig, err := dbgen.FindConfigurationG(ctx, configID)
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

func GetTenantConfig(ctx context.Context, configID int64, tenantId uuid.UUID) (appmodel.Configuration, error) {
	dbConfig, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(configID),
		dbgen.ConfigurationWhere.TenantID.EQ(tenantId.String()),
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

func GetTenantConfigs(ctx context.Context, tenantId uuid.UUID) ([]appmodel.Configuration, error) {
	dbConfigs, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.TenantID.EQ(tenantId.String()),
	).AllG(ctx)
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

func DeleteConfig(ctx context.Context, configID int64, tenantId uuid.UUID) error {
	// First, verify that the configuration exists and belongs to the specified tenant
	configExists, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(configID),
		dbgen.ConfigurationWhere.TenantID.EQ(tenantId.String()),
	).ExistsG(ctx)
	if err != nil {
		return fmt.Errorf("checking config existence: %v", err)
	}
	if !configExists {
		return ErrNotFound
	}

	// Delete assets that belong to this configuration
	if _, err := dbgen.Assets(
		dbgen.AssetWhere.ConfigurationID.EQ(configID),
	).DeleteAllG(ctx); err != nil {
		return fmt.Errorf("deleting assets from database: %v", err)
	}

	// Delete the configuration that matches both configID and tenantId
	count, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(configID),
		dbgen.ConfigurationWhere.TenantID.EQ(tenantId.String()),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("deleting config from database: %v", err)
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) configs by ID and tenant", count)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func SetConfigActiveState(ctx context.Context, config appmodel.Configuration, state bool) (int64, error) {
	return dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(config.Id),
	).UpdateAllG(ctx, dbgen.M{
		dbgen.ConfigurationColumns.Active: state,
	})
}

func InsertAsset(ctx context.Context, asset appmodel.Asset) error {
	stmt := Asset.INSERT(
		Asset.ProjectID,
		Asset.AssetID,
		Asset.LocationName,
		Asset.Lat,
		Asset.Lon,
	).VALUES(
		asset.ProjectID,
		asset.AssetID,
		asset.LocationName,
		asset.Lat,
		asset.Lon,
	).ON_CONFLICT(
		Asset.AssetID,
	).DO_NOTHING()

	_, err := stmt.ExecContext(ctx, GetDB().db)
	return err
}

// rework the old to new
func InsertAsset(ctx context.Context, config appmodel.Configuration, globalAssetID string, assetId int32, providerId string, isRoot bool) error {
	dbAsset := dbgen.Asset{
		ConfigurationID: config.Id,
		GlobalAssetID:   globalAssetID,
		AssetID:         null.Int32From(assetId),
		ProviderID:      providerId,
		IsRoot:          isRoot,
	}
	return dbAsset.UpsertG(ctx, true, []string{dbgen.AssetColumns.ProviderID}, boil.Blacklist("id"), boil.Infer())
}

func UpdateAssetLocation(ctx context.Context, asset appmodel.Asset) error {
	stmt := Asset.UPDATE(
		Asset.LocationName,
		Asset.Lat,
		Asset.Lon,
	).SET(
		asset.LocationName,
		asset.Lat,
		asset.Lon,
	).WHERE(
		Asset.ID.EQ(Int(asset.ID)),
	)
	_, err := stmt.ExecContext(ctx, GetDB().db)
	return err
}

// rework old to new
func UpdateAssetLocation(ctx context.Context, config appmodel.Configuration, globalAssetID string, assetId int32, providerId string, isRoot bool) error {
	dbAsset := dbgen.Asset{
		ConfigurationID: config.Id,
		GlobalAssetID:   globalAssetID,
		AssetID:         null.Int32From(assetId),
		ProviderID:      providerId,
		IsRoot:          isRoot,
	}
	return dbAsset.UpsertG(ctx, true, []string{dbgen.AssetColumns.ProviderID}, boil.Blacklist("id"), boil.Infer())
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

func GetAssetById(assetId int32) (appmodel.Asset, error) {
	//Not used anywhere, but probably will need to support tenantId if used in future
	asset, err := dbgen.FindAssetG(context.Background(), int64(assetId))
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}
	if !asset.AssetID.Valid {
		return appmodel.Asset{}, fmt.Errorf("shouldn't happen: assetID is nil")
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

func GetAssets(ctx context.Context) ([]appmodel.Asset, error) {
	var assets []model.Asset
	err := SELECT(
		Asset.AllColumns,
	).FROM(
		Asset,
	).QueryContext(ctx, GetDB().db, &assets)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("fetching assets: %v", err)
	}

	var appAssets []appmodel.Asset
	for _, a := range assets {
		appAssets = append(appAssets, toAppAsset(a))
	}
	return appAssets, nil

}

func toDbConfig(ctx context.Context, appConfig appmodel.Configuration) (dbConfig dbgen.Configuration, err error) {
	dbConfig.TenantID = appConfig.TenantId.String()

	dbConfig.ID = appConfig.Id
	dbConfig.SiteID = null.StringFrom(appConfig.SiteID)
	dbConfig.RefreshInterval = appConfig.RefreshInterval
	dbConfig.RequestTimeout = appConfig.RequestTimeout
	af, err := json.Marshal(appConfig.AssetFilter)
	if err != nil {
		return dbgen.Configuration{}, fmt.Errorf("marshalling assetFilter: %v", err)
	}
	dbConfig.AssetFilter = af
	dbConfig.Active = appConfig.Active
	dbConfig.Enable = appConfig.Enable

	env := frontend.GetEnvironment(ctx)
	if env != nil {
		dbConfig.UserID = env.UserId
	}

	return dbConfig, nil
}

func toAppConfig(dbConfig *dbgen.Configuration) (appConfig appmodel.Configuration, err error) {
	var apikey_err error

	tenantUUID, err := uuid.Parse(dbConfig.TenantID)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("parsing tenant ID as UUID: %v", err)
	}

	appConfig.ApiKey, apikey_err = app.GetApiKey("demo", GetDB(), tenantUUID)
	if apikey_err != nil {
		log.Fatal("conf", "api key not found for tenant %s in DB for: %v", dbConfig.TenantID, apikey_err)
	}

	appConfig.Id = dbConfig.ID
	appConfig.TenantId = tenantUUID
	appConfig.SiteID = dbConfig.SiteID.String
	appConfig.Enable = dbConfig.Enable
	appConfig.RefreshInterval = dbConfig.RefreshInterval
	appConfig.RequestTimeout = dbConfig.RequestTimeout
	var af [][]appmodel.FilterRule
	if err := json.Unmarshal(dbConfig.AssetFilter, &af); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("unmarshalling assetFilter: %v", err)
	}
	appConfig.AssetFilter = af
	appConfig.Active = dbConfig.Active
	appConfig.UserId = dbConfig.UserID
	return appConfig, nil
}

func toAppAsset(dbAsset dbgen.Asset, config appmodel.Configuration) appmodel.Asset {
	return appmodel.Asset{
		ID:            dbAsset.ID,
		Config:        config,
		ProjectID:     dbAsset.ProjectID,
		GlobalAssetID: dbAsset.GlobalAssetID,
		ProviderID:    dbAsset.ProviderID,
		AssetID:       dbAsset.AssetID.Int32,
	}
}

func UpsertRootAsset(assetID int32, projectID, gai string) error {
	stmt := RootAsset.INSERT(
		RootAsset.ConfigurationID,
		RootAsset.Gai,
		RootAsset.ProjectID,
		RootAsset.AssetID,
	).VALUES(
		1,
		gai,
		projectID,
		assetID,
	).ON_CONFLICT(
		RootAsset.AssetID,
	).DO_NOTHING()

	_, err := stmt.ExecContext(context.Background(), GetDB().db)
	return err
}

func GetRootAssets() ([]appmodel.RootAsset, error) {
	var assets []model.Asset
	err := SELECT(
		RootAsset.AllColumns,
	).FROM(
		RootAsset,
	).Query(GetDB().db, &assets)
	if err != nil {
		return nil, fmt.Errorf("fetching root assets: %v", err)
	}

	appAssets := make([]appmodel.RootAsset, 0, len(assets))
	for _, asset := range assets {
		appAssets = append(appAssets, appmodel.RootAsset{
			ID:      asset.ID,
			AssetID: asset.AssetID,
		})
	}
	return appAssets, nil
}

func GetRootAssets(tenantId uuid.UUID) ([]appmodel.Asset, error) {
	// rework because weather-app2 uses separete db table for root asset
	assets, err := dbgen.Assets(
		dbgen.AssetWhere.IsRoot.EQ(true),
	).AllG(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching root assets: %v", err)
	}

	return FilterAssetsByTenant(assets, tenantId)
}

func GetRootAssetId(ctx context.Context, projectID, gai string) (*int32, error) {
	var dest struct {
		ID int32
	}
	stmt := RootAsset.SELECT(
		RootAsset.ID,
	).WHERE(
		RootAsset.Gai.EQ(String(gai)).AND(
			RootAsset.ProjectID.EQ(String(projectID)),
		),
	)
	err := stmt.QueryContext(ctx, GetDB().db, &dest)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("getting root asset ID: %v", err)
	}

	return &dest.ID, nil
}

func RootAssetAlreadyCreated() (bool, error) {
	var dest struct {
		ID int32
	}
	stmt := RootAsset.SELECT(
		RootAsset.ID,
	)
	err := stmt.QueryContext(context.Background(), GetDB().db, &dest)
	if errors.Is(err, qrm.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("getting root asset: %v", err)
	}

	return true, nil
}

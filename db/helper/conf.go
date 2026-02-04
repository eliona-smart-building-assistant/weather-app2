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
	"github.com/google/uuid"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

func UpsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	// we assume unique 1 config per 1 tenant
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}
	if err := dbConfig.UpsertG(ctx, true, []string{"tenant_id"}, boil.Blacklist("id"), boil.Infer()); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}
	return config, nil
}

func GetTenantConfig(ctx context.Context, tenantId uuid.UUID) (appmodel.Configuration, error) {
	dbConfig, err := dbgen.Configurations(
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

func SetConfigActiveState(ctx context.Context, config appmodel.Configuration, state bool) (int64, error) {
	return dbgen.Configurations(
		dbgen.ConfigurationWhere.ID.EQ(config.Id),
	).UpdateAllG(ctx, dbgen.M{
		dbgen.ConfigurationColumns.Active: state,
	})
}

func InsertAsset(ctx context.Context, asset appmodel.Asset) error {
	dbAsset := dbgen.Asset{
		ProjectID:    asset.ProjectID,
		LocationName: asset.LocationName,
		Lat:          asset.Lat,
		Lon:          asset.Lon,
		AssetID:      asset.AssetID,
	}
	return dbAsset.UpsertG(ctx, true, []string{dbgen.AssetColumns.AssetID}, boil.Blacklist("id"), boil.Infer())
}

func UpdateAssetLocation(ctx context.Context, asset appmodel.Asset) error {
	count, err := dbgen.Assets(
		dbgen.AssetWhere.ID.EQ(asset.ID),
	).UpdateAllG(ctx, dbgen.M{
		dbgen.AssetColumns.LocationName: asset.LocationName,
		dbgen.AssetColumns.Lat:          asset.Lat,
		dbgen.AssetColumns.Lon:          asset.Lon,
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no asset found with ID %d", asset.ID)
	}
	return nil
}

func GetAssetById(assetId int32) (appmodel.Asset, error) {
	asset, err := dbgen.FindAssetG(context.Background(), int64(assetId))
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}

	return toAppAsset(*asset), nil
}

func GetAssets(ctx context.Context) ([]appmodel.Asset, error) {
	assets, err := dbgen.Assets().AllG(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching assets: %v", err)
	}
	var appAssets []appmodel.Asset
	for _, a := range assets {
		appAssets = append(appAssets, toAppAsset(*a))
	}
	return appAssets, nil
}

func UpsertRootAsset(ctx context.Context, rootAsset appmodel.RootAsset) error {
	// we assume unique 1 config per 1 tenant
	dbRootAsset := toDbRootAsset(rootAsset)

	if err := dbRootAsset.UpsertG(ctx, true, []string{"AssetID"}, boil.Blacklist("id"), boil.Infer()); err != nil {
		return fmt.Errorf("inserting DB rootAsset: %v", err)
	}
	return nil
}

func GetRootAssets() ([]appmodel.RootAsset, error) {
	rootAssets, err := dbgen.RootAssets().AllG(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching root assets: %v", err)
	}
	var appRootAssets []appmodel.RootAsset
	for _, ra := range rootAssets {
		appRootAssets = append(appRootAssets, toAppRootAsset(*ra))
	}

	return appRootAssets, nil
}

func GetRootAssetId(ctx context.Context, globalAssetID string, config appmodel.Configuration) (*int32, error) {
	dbRootAsset, err := dbgen.RootAssets(
		dbgen.RootAssetWhere.Gai.EQ(globalAssetID),
		dbgen.RootAssetWhere.ConfigurationID.EQ(config.Id),
	).OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting root asset: %v", err)
	}
	return common.Ptr(dbRootAsset.AssetID), nil
}

func RootAssetAlreadyCreated(tenantId uuid.UUID) (bool, error) {
	// First get the configuration for this tenant
	config, err := dbgen.Configurations(
		dbgen.ConfigurationWhere.TenantID.EQ(tenantId.String()),
	).OneG(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		// No configuration exists for this tenant, so no root assets either
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("getting configuration for tenant %s: %v", tenantId, err)
	}

	// Now check if root assets exist for this configuration
	exists, err := dbgen.RootAssets(
		dbgen.RootAssetWhere.ConfigurationID.EQ(config.ID),
	).ExistsG(context.Background())
	if err != nil {
		return false, fmt.Errorf("checking if root asset exists for tenant %s: %v", tenantId, err)
	}
	return exists, nil
}

func toDbConfig(ctx context.Context, appConfig appmodel.Configuration) (dbConfig dbgen.Configuration, err error) {
	dbConfig.TenantID = appConfig.TenantId.String()

	dbConfig.ID = appConfig.Id
	dbConfig.SiteID = null.StringFrom(appConfig.SiteId)
	dbConfig.RefreshInterval = appConfig.RefreshInterval
	dbConfig.RequestTimeout = appConfig.RequestTimeout
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

	appConfig.ApiKey, apikey_err = app.GetApiKey("weather-app2", GetDB(), tenantUUID)
	if apikey_err != nil {
		log.Fatal("conf", "api key not found for tenant %s in DB for: %v", dbConfig.TenantID, apikey_err)
	}

	appConfig.Id = dbConfig.ID
	appConfig.TenantId = tenantUUID
	appConfig.SiteId = dbConfig.SiteID.String
	appConfig.Enable = dbConfig.Enable
	appConfig.RefreshInterval = dbConfig.RefreshInterval
	appConfig.RequestTimeout = dbConfig.RequestTimeout
	appConfig.Active = dbConfig.Active
	appConfig.UserId = dbConfig.UserID
	return appConfig, nil
}

func toAppAsset(dbAsset dbgen.Asset) appmodel.Asset {
	return appmodel.Asset{
		ID:           dbAsset.ID,
		ProjectID:    dbAsset.ProjectID,
		LocationName: dbAsset.LocationName,
		Lat:          dbAsset.Lat,
		Lon:          dbAsset.Lon,
		AssetID:      dbAsset.AssetID,
	}
}

func toAppRootAsset(dbRootAsset dbgen.RootAsset) appmodel.RootAsset {
	return appmodel.RootAsset{
		ID: dbRootAsset.ID,
		// config: 	  dbRootAsset.ConfigurationID, // Probably don't need
		AssetID: dbRootAsset.AssetID,
	}
}

func toDbRootAsset(appRootAsset appmodel.RootAsset) (dbRootAsset dbgen.RootAsset) {
	dbRootAsset.ID = appRootAsset.ID
	dbRootAsset.ConfigurationID = appRootAsset.Config.Id
	dbRootAsset.Gai = appRootAsset.Gai
	dbRootAsset.AssetID = appRootAsset.AssetID

	return dbRootAsset
}

func ParseTenantIdFromEnv(ctx context.Context) (uuid.UUID, error) {
	env := frontend.GetEnvironment(ctx)
	if env == nil {
		return uuid.UUID{}, fmt.Errorf("missing environment JWT")
	}
	parsed, err := uuid.Parse(env.TenantId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("tenant isn't a valid UUID: %s", env.TenantId)
	}
	return parsed, err
}

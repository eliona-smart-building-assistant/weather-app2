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
	appmodel "weather-app2/app/model"

	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/lib/pq"

	"weather-app2/db/generated/postgres/weather_app/model"
	. "weather-app2/db/generated/postgres/weather_app/table"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

// DBHelper is a singleton struct managing the database connection and queries.
type DBHelper struct {
	db *sql.DB
}

var (
	instance *DBHelper
)

// InitDB initializes the database connection ONCE.
func InitDB(db *sql.DB) {
	instance = &DBHelper{
		db: db,
	}
}

// GetDB returns the singleton database instance.
func GetDB() *DBHelper {
	if instance == nil {
		log.Fatal("conf", "Database not initialized. Call InitDB() first.")
	}
	return instance
}

// CloseDB gracefully shuts down the database connection.
func CloseDB() error {
	if instance != nil && instance.db != nil {
		return instance.db.Close()
	}
	return nil
}

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

func UpsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	commonColumns := ColumnList{
		Configuration.APIKey,
		Configuration.RefreshInterval,
		Configuration.RequestTimeout,
		Configuration.Active,
		Configuration.Enable,
		Configuration.ProjectIds,
		Configuration.UserID,
	}

	commonValues := []interface{}{
		config.ApiKey,
		config.RefreshInterval,
		config.RequestTimeout,
		config.Active,
		config.Enable,
		pq.StringArray(config.ProjectIDs),
		frontend.GetEnvironment(ctx).UserId,
	}

	stmt := Configuration.INSERT()

	if config.Id != 0 {
		// If ID is provided, include it in the INSERT
		columns := append(commonColumns, Configuration.ID)
		values := append(commonValues, config.Id)

		stmt = Configuration.INSERT(columns).VALUES(values[0], values[1:]...).ON_CONFLICT(
			Configuration.ID,
		).DO_UPDATE(
			SET(
				Configuration.APIKey.SET(Configuration.EXCLUDED.APIKey),
				Configuration.RefreshInterval.SET(Configuration.EXCLUDED.RefreshInterval),
				Configuration.RequestTimeout.SET(Configuration.EXCLUDED.RequestTimeout),
				Configuration.Active.SET(Configuration.EXCLUDED.Active),
				Configuration.Enable.SET(Configuration.EXCLUDED.Enable),
				Configuration.ProjectIds.SET(Configuration.EXCLUDED.ProjectIds),
			),
		)
	} else {
		// If ID is 0, omit it to allow auto-increment
		stmt = Configuration.INSERT(commonColumns).VALUES(commonValues[0], commonValues[1:]...)
	}

	stmt = stmt.RETURNING(Configuration.AllColumns)

	var updatedConfig model.Configuration
	if err := stmt.QueryContext(ctx, GetDB().db, &updatedConfig); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("upserting config: %v", err)
	}

	return toAppConfig(updatedConfig)
}

func GetConfig(ctx context.Context) (appmodel.Configuration, error) {
	var dbConfig model.Configuration
	err := Configuration.
		SELECT(Configuration.AllColumns).
		QueryContext(ctx, GetDB().db, &dbConfig)
	if errors.Is(err, qrm.ErrNoRows) {
		return appmodel.Configuration{}, ErrNotFound
	} else if err != nil {
		return appmodel.Configuration{}, err
	}

	return toAppConfig(dbConfig)
}

func SetConfigActiveState(ctx context.Context, state bool) error {
	stmt := Configuration.UPDATE(Configuration.Active).
		SET(state)
	_, err := stmt.ExecContext(ctx, GetDB().db)
	return err
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

func GetAssetId(ctx context.Context, config appmodel.Configuration, projectID, assetID int32) (*int32, error) {
	var dest struct {
		ID int32
	}
	stmt := Asset.SELECT(
		Asset.ID,
	).WHERE(
		Asset.AssetID.EQ(Int(int64(assetID))),
	)
	err := stmt.QueryContext(ctx, GetDB().db, &dest)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("getting asset ID: %v", err)
	}

	return &dest.ID, nil
}

type assetWithConfig struct {
	model.Asset
	model.Configuration
}

func GetAssetById(assetId int32) (appmodel.Asset, error) {
	var asset assetWithConfig
	err := SELECT(
		Asset.AllColumns,
	).FROM(
		Asset,
	).WHERE(
		Asset.AssetID.EQ(Int32(assetId)),
	).Query(GetDB().db, &asset)
	if errors.Is(err, qrm.ErrNoRows) {
		return appmodel.Asset{}, ErrNotFound
	} else if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching asset %v: %v", assetId, err)
	}

	return toAppAsset(asset.Asset), nil
}

func toAppConfig(dbCfg model.Configuration) (appmodel.Configuration, error) {
	return appmodel.Configuration{
		Id:              1,
		ApiKey:          dbCfg.APIKey,
		RefreshInterval: dbCfg.RefreshInterval,
		RequestTimeout:  dbCfg.RequestTimeout,
		Active:          dbCfg.Active,
		Enable:          dbCfg.Enable,
		ProjectIDs:      dbCfg.ProjectIds,
		UserId:          dbCfg.UserID,
	}, nil
}

func toAppAsset(dbAsset model.Asset) appmodel.Asset {
	return appmodel.Asset{
		ID:           dbAsset.ID,
		ProjectID:    dbAsset.ProjectID,
		LocationName: dbAsset.LocationName,
		Lat:          dbAsset.Lat,
		Lon:          dbAsset.Lon,
		AssetID:      dbAsset.AssetID,
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

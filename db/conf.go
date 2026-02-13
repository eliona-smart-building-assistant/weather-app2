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

	"github.com/eliona-smart-building-assistant/go-eliona/v2/app"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/google/uuid"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

// UpsertConfig inserts or updates a configuration
func UpsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	env := frontend.GetEnvironment(ctx)
	userID := ""
	if env != nil {
		userID = env.UserId
	}

	query := `
		INSERT INTO configuration (tenant_id, site_id, refresh_interval, request_timeout, active, enable, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(tenant_id) DO UPDATE SET
			site_id = excluded.site_id,
			refresh_interval = excluded.refresh_interval,
			request_timeout = excluded.request_timeout,
			active = excluded.active,
			enable = excluded.enable,
			user_id = excluded.user_id
		RETURNING id`

	var id int64
	err := GetDB().QueryRowContext(ctx, query,
		config.TenantId.String(),
		sql.NullString{String: config.SiteId, Valid: config.SiteId != ""},
		config.RefreshInterval,
		config.RequestTimeout,
		config.Active,
		config.Enable,
		userID,
	).Scan(&id)

	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("upserting config: %v", err)
	}

	config.Id = id
	return config, nil
}

// GetTenantConfig retrieves a configuration by tenant ID
func GetTenantConfig(ctx context.Context, tenantId uuid.UUID) (appmodel.Configuration, error) {
	query := `SELECT id, tenant_id, site_id, refresh_interval, request_timeout, active, enable, user_id
	          FROM configuration WHERE tenant_id = ?`

	var config appmodel.Configuration
	var siteID sql.NullString
	var tenantIDStr string

	err := GetDB().QueryRowContext(ctx, query, tenantId.String()).Scan(
		&config.Id,
		&tenantIDStr,
		&siteID,
		&config.RefreshInterval,
		&config.RequestTimeout,
		&config.Active,
		&config.Enable,
		&config.UserId,
	)

	if err == sql.ErrNoRows {
		return appmodel.Configuration{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("fetching config: %v", err)
	}

	config.TenantId = tenantId
	if siteID.Valid {
		config.SiteId = siteID.String
	}

	// Get API key
	apiKey, err := app.GetApiKey("demo", GetDB(), tenantId)
	if err != nil {
		log.Error("conf", "api key not found for tenant %s: %v", tenantId, err)
	}
	config.ApiKey = apiKey

	return config, nil
}

// GetConfigs retrieves all configurations
func GetConfigs(ctx context.Context) ([]appmodel.Configuration, error) {
	query := `SELECT id, tenant_id, site_id, refresh_interval, request_timeout, active, enable, user_id
	          FROM configuration`

	rows, err := GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying configs: %v", err)
	}
	defer rows.Close()

	var configs []appmodel.Configuration
	for rows.Next() {
		var config appmodel.Configuration
		var siteID sql.NullString
		var tenantIDStr string

		err := rows.Scan(
			&config.Id,
			&tenantIDStr,
			&siteID,
			&config.RefreshInterval,
			&config.RequestTimeout,
			&config.Active,
			&config.Enable,
			&config.UserId,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning config: %v", err)
		}

		tenantUUID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("parsing tenant ID: %v", err)
		}
		config.TenantId = tenantUUID

		if siteID.Valid {
			config.SiteId = siteID.String
		}

		// Get API key
		apiKey, err := app.GetApiKey("demo", GetDB(), tenantUUID)
		if err != nil {
			log.Error("conf", "api key not found for tenant %s: %v", tenantUUID, err)
		}
		config.ApiKey = apiKey

		configs = append(configs, config)
	}

	return configs, rows.Err()
}

// SetConfigActiveState updates the active state of a configuration
func SetConfigActiveState(ctx context.Context, config appmodel.Configuration, state bool) (int64, error) {
	query := `UPDATE configuration SET active = ? WHERE id = ?`
	result, err := GetDB().ExecContext(ctx, query, state, config.Id)
	if err != nil {
		return 0, fmt.Errorf("updating active state: %v", err)
	}
	return result.RowsAffected()
}

// InsertAsset inserts
func InsertAsset(ctx context.Context, asset appmodel.Asset) error {
	query := `
		INSERT INTO asset (location_name, lat, lon, asset_id)
		VALUES (?, ?, ?, ?)`

	_, err := GetDB().ExecContext(ctx, query, asset.LocationName, asset.Lat, asset.Lon, asset.AssetID)
	if err != nil {
		return fmt.Errorf("inserting asset: %v", err)
	}
	return nil
}

// UpdateAssetLocation updates location fields of an asset
func UpdateAssetLocation(ctx context.Context, asset appmodel.Asset) error {
	query := `UPDATE asset SET location_name = ?, lat = ?, lon = ? WHERE id = ?`
	result, err := GetDB().ExecContext(ctx, query, asset.LocationName, asset.Lat, asset.Lon, asset.ID)
	if err != nil {
		return fmt.Errorf("updating asset location: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no asset found with ID %d", asset.ID)
	}
	return nil
}

// GetAssetId retrieves asset ID by configuration and global asset ID
func GetAssetId(ctx context.Context, config appmodel.Configuration, globalAssetID string) (*int32, error) {
	query := `SELECT asset_id FROM asset WHERE configuration_id = ? AND global_asset_id = ?`

	var assetID sql.NullInt32
	err := GetDB().QueryRowContext(ctx, query, config.Id, globalAssetID).Scan(&assetID)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetching asset ID: %v", err)
	}

	if !assetID.Valid {
		return nil, nil
	}

	return common.Ptr(assetID.Int32), nil
}

// GetAssetById retrieves an asset by its ID
func GetAssetById(assetId int32) (appmodel.Asset, error) {
	query := `SELECT a.id, a.configuration_id, a.location_name, a.lat, a.lon, a.asset_id
	          FROM asset a
	          WHERE a.asset_id = ?`

	var asset appmodel.Asset

	err := GetDB().QueryRow(query, assetId).Scan(
		&asset.ID,
		&asset.ConfigurationId,
		&asset.LocationName,
		&asset.Lat,
		&asset.Lon,
		&asset.AssetID,
	)

	if err == sql.ErrNoRows {
		return appmodel.Asset{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}

	return asset, nil
}

// GetAssets retrieves all assets
func GetAssets(ctx context.Context) ([]appmodel.Asset, error) {
	query := `SELECT id, configuration_id, location_name, lat, lon, asset_id
	          FROM asset`

	rows, err := GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying assets: %v", err)
	}
	defer rows.Close()

	var assets []appmodel.Asset
	for rows.Next() {
		var asset appmodel.Asset

		err := rows.Scan(
			&asset.ID,
			&asset.ConfigurationId,
			&asset.LocationName,
			&asset.Lat,
			&asset.Lon,
			&asset.AssetID,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning asset: %v", err)
		}

		assets = append(assets, asset)
	}

	return assets, rows.Err()
}

// UpsertRootAsset inserts or updates a root asset
func UpsertRootAsset(ctx context.Context, rootAsset appmodel.RootAsset) error {
	query := `
		INSERT INTO root_asset (configuration_id, project_id, gai, asset_id)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(asset_id) DO UPDATE SET
			configuration_id = excluded.configuration_id,
			project_id = excluded.project_id,
			gai = excluded.gai`

	_, err := GetDB().ExecContext(ctx, query,
		rootAsset.Config.Id,
		"",
		rootAsset.Gai,
		rootAsset.AssetID,
	)

	if err != nil {
		return fmt.Errorf("upserting root asset: %v", err)
	}
	return nil
}

// GetRootAssets retrieves all root assets
func GetRootAssets() ([]appmodel.RootAsset, error) {
	query := `SELECT id, configuration_id, gai, asset_id FROM root_asset`

	rows, err := GetDB().Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying root assets: %v", err)
	}
	defer rows.Close()

	var rootAssets []appmodel.RootAsset
	for rows.Next() {
		var ra appmodel.RootAsset
		err := rows.Scan(&ra.ID, &ra.Config.Id, &ra.Gai, &ra.AssetID)
		if err != nil {
			return nil, fmt.Errorf("scanning root asset: %v", err)
		}
		rootAssets = append(rootAssets, ra)
	}

	return rootAssets, rows.Err()
}

// GetRootAssetId retrieves root asset ID by global asset ID and configuration
func GetRootAssetId(ctx context.Context, globalAssetID string, config appmodel.Configuration) (*int32, error) {
	query := `SELECT asset_id FROM root_asset WHERE gai = ? AND configuration_id = ?`

	var assetID int32
	err := GetDB().QueryRowContext(ctx, query, globalAssetID, config.Id).Scan(&assetID)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetching root asset ID: %v", err)
	}

	return common.Ptr(assetID), nil
}

// RootAssetAlreadyCreated checks if a root asset exists for the given tenant
func RootAssetAlreadyCreated(tenantId uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM root_asset ra
			JOIN configuration c ON ra.configuration_id = c.id
			WHERE c.tenant_id = ?
		)`

	var exists bool
	err := GetDB().QueryRow(query, tenantId.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking root asset existence: %v", err)
	}

	return exists, nil
}

// ParseTenantIdFromEnv parses tenant ID from environment/context
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

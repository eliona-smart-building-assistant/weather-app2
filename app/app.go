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
	"net/http"
	"reflect"
	"sync"
	"time"
	apiserver "weather-app2/api/generated"
	apiservices "weather-app2/api/services"
	appmodel "weather-app2/app/model"
	dbhelper "weather-app2/db/helper"
	"weather-app2/eliona"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/dashboard"
	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
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

func Initialize() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// Init the app before the first run.
	app.Init(conn, app.AppName(),
		app.ExecSqlFile("db/init.sql"),
		asset.InitAssetTypeFiles("resources/asset-types/*.json"),
		dashboard.InitWidgetTypeFiles("resources/widget-types/*.json"),
	)
}

var (
	once             sync.Once
	configChangeChan = make(chan struct{})
	previousConfigs  = make(map[int64]appmodel.Configuration)
	configMutex      sync.Mutex
)

func CollectData() {
	configs, err := dbhelper.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("dbhelper", "Couldn't read configs from DB: %v", err)
		changeAppStatus(statusFatal)
		return
	}
	if len(configs) == 0 {
		once.Do(func() {
			log.Info("dbhelper", "No configs in DB. Please configure the app in Eliona.")
		})
		return
	}

	for _, config := range configs {
		if !config.Enable {
			if config.Active {
				dbhelper.SetConfigActiveState(context.Background(), config.Id, false)
			}
			continue
		}

		if !config.Active {
			dbhelper.SetConfigActiveState(context.Background(), config.Id, true)
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

		// Check for changes in this specific config
		if isConfigChanged(config) {
			configChangeChan <- struct{}{}
		}

		ctx, cancel := context.WithCancel(context.Background())
		common.RunOnceWithParam(func(config appmodel.Configuration) {
			log.Info("main", "Collecting %d started.", config.Id)
			if err := collectResources(ctx, &config); err != nil {
				changeAppStatus(statusError)
				cancel() // Cancel the context to stop the long-running processes
				return   // Error is handled in the method itself.
			}
			log.Info("main", "Collecting %d finished.", config.Id)
			changeAppStatus(statusOK)

			// Wait for the next interval or a config change
			select {
			case <-time.After(time.Second * time.Duration(config.RefreshInterval)):
				// Continue with the next iteration
				return
			case <-configChangeChan:
				// Config changed, restart the process
				cancel() // Cancel the context to stop the long-running process
				return
			}
		}, config, config.Id)
	}
}

func isConfigChanged(newConfig appmodel.Configuration) bool {
	configMutex.Lock()
	defer configMutex.Unlock()

	oldConfig, exists := previousConfigs[newConfig.Id]
	if !exists {
		// New config added
		previousConfigs[newConfig.Id] = newConfig
		return true
	}

	if !reflect.DeepEqual(newConfig, oldConfig) {
		// Config changed
		previousConfigs[newConfig.Id] = newConfig
		return true
	}

	return false
}

func collectResources(ctx context.Context, config *appmodel.Configuration) error {
	// Do the magic here
	return nil
}

// ListenForOutputChanges listens to output attribute changes from Eliona. Delete if not needed.
func ListenForOutputChanges() {
	for { // We want to restart listening in case something breaks.
		outputs, err := eliona.ListenForOutputChanges()
		if err != nil {
			log.Error("eliona", "listening for output changes: %v", err)
			changeAppStatus(statusError)
			return
		}
		for output := range outputs {
			if cr := output.ClientReference.Get(); cr != nil && *cr == eliona.ClientReference {
				// Just an echoed value this app sent.
				continue
			}
			asset, err := dbhelper.GetAssetById(output.AssetId)
			if errors.Is(err, dbhelper.ErrNotFound) {
				log.Debug("app", "received data update for other apps asset %v", output.AssetId)
				continue
			} else if err != nil {
				log.Error("dbhelper", "getting asset by assetID %v: %v", output.AssetId, err)
				changeAppStatus(statusError)
				return
			}
			if err := outputData(asset, output.Data); err != nil {
				log.Error("dbhelper", "outputting data (%v) for config %v and assetId %v: %v", output.Data, asset.Config.Id, asset.AssetID, err)
				changeAppStatus(statusError)
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

func Heartbeat() {
	roots, err := dbhelper.GetRootAssets()
	if err != nil {
		log.Error("dbhelper", "getting root assets: %v", err)
		return
	}

	for _, root := range roots {
		err := eliona.UpsertData(root.AssetID, map[string]any{"status": appStatus}, time.Now(), api.SUBTYPE_STATUS)
		if err != nil {
			log.Error("eliona", "upserting data as heartbeat: %v", err)
			return
		}
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

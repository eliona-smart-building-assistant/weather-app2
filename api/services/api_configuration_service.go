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

package apiservices

import (
	"context"
	"fmt"
	"net/http"
	apiserver "weather-app2/api/generated"
	appmodel "weather-app2/app/model"
	"weather-app2/broker"
	dbhelper "weather-app2/db/helper"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
)

// ConfigurationAPIService is a service that implements the logic for the ConfigurationAPIServicer
// This service should implement the business logic for every endpoint for the ConfigurationAPI API.
// Include any external packages or services that will be required by this service.
type ConfigurationAPIService struct {
}

// NewConfigurationAPIService creates a default api service
func NewConfigurationAPIService() apiserver.ConfigurationAPIServicer {
	return &ConfigurationAPIService{}
}

func (s *ConfigurationAPIService) GetConfiguration(ctx context.Context) (apiserver.ImplResponse, error) {
	appConfig, err := dbhelper.GetConfig(ctx)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusOK, toAPIConfig(appConfig)), nil
}

func (s *ConfigurationAPIService) PutConfiguration(ctx context.Context, config apiserver.Configuration) (apiserver.ImplResponse, error) {
	config.Id = api.PtrInt64(1)
	appConfig := toAppConfig(config)
	if err := broker.TestAuthentication(appConfig); err != nil {
		return apiserver.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("testing authentication: %v", err)
	}
	upsertedConfig, err := dbhelper.UpsertConfig(ctx, appConfig)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusCreated, toAPIConfig(upsertedConfig)), nil
}

func toAPIConfig(appConfig appmodel.Configuration) apiserver.Configuration {
	return apiserver.Configuration{
		Id:              &appConfig.Id,
		ApiKey:          appConfig.ApiKey,
		Enable:          &appConfig.Enable,
		RefreshInterval: appConfig.RefreshInterval,
		RequestTimeout:  &appConfig.RequestTimeout,
		Active:          &appConfig.Active,
		ProjectIDs:      &appConfig.ProjectIDs,
		UserId:          &appConfig.UserId,
	}
}

func toAppConfig(apiConfig apiserver.Configuration) (appConfig appmodel.Configuration) {
	appConfig.ApiKey = apiConfig.ApiKey

	if apiConfig.Id != nil {
		appConfig.Id = *apiConfig.Id
	}
	appConfig.RefreshInterval = apiConfig.RefreshInterval
	if apiConfig.RequestTimeout != nil {
		appConfig.RequestTimeout = *apiConfig.RequestTimeout
	}

	if apiConfig.Active != nil {
		appConfig.Active = *apiConfig.Active
	}
	if apiConfig.Enable != nil {
		appConfig.Enable = *apiConfig.Enable
	}
	if apiConfig.ProjectIDs != nil {
		appConfig.ProjectIDs = *apiConfig.ProjectIDs
	}
	return appConfig
}

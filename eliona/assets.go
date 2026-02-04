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

package eliona

import (
	"fmt"
	appmodel "weather-app2/v2/app/model"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v3"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/client"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

var devicesCount map[int64]int

func CreateAssets(config appmodel.Configuration, assets []asset.AssetLikeWithParentReferences) error {
	// TODO: remove this workaround once the assetsCreated is returned correctly again
	if devicesCount == nil {
		devicesCount = make(map[int64]int)
	}
	// TODO: this does not return assets created anymore, but total number of assets!
	assetsCreated, err := asset.CreateAssetsBulk(client.ApiEndpointString(), config.ApiKey, assets)
	if err != nil {
		return err
	}
	log.Debug("eliona", "finished creating %v assets", assetsCreated)
	if assetsCreated != 0 && devicesCount[config.Id] != assetsCreated {
		if err := notifyUser(config.UserId, config.SiteId, assetsCreated, config); err != nil {
			return fmt.Errorf("notifying user about CAC: %v", err)
		}
		devicesCount[config.Id] = assetsCreated
	}
	return nil
}

func notifyUser(userId string, siteId string, assetsCreated int, config appmodel.Configuration) error {
	receipt, _, err := client.NewClient(client.ApiEndpointString()).CommunicationAPI.
		PostNotification(client.AuthenticationContext(config.ApiKey)).
		Notification(
			api.Notification{
				User:   userId,
				SiteId: *api.NewNullableString(&siteId),
				Message: *api.NewNullableTranslation(&api.Translation{
					De: api.PtrString(fmt.Sprintf("Weather App hat %d neue Assets angelegt. Diese sind nun im Asset-Management verfügbar.", assetsCreated)),
					En: api.PtrString(fmt.Sprintf("Wetter app added %v new assets. They are now available in Asset Management.", assetsCreated)),
				}),
			}).
		Execute()
	log.Debug("eliona", "posted notification about CAC: %v", receipt)
	if err != nil {
		return fmt.Errorf("posting CAC notification: %v", err)
	}
	return nil
}

func GetAsset(assetID int32, config appmodel.Configuration) (*api.Asset, error) {
	asset, _, err := client.NewClient(client.ApiEndpointString()).AssetsAPI.GetAssetById(client.AuthenticationContext(config.ApiKey), assetID).Execute()
	return asset, err
}

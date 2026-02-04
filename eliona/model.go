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
	"context"
	appmodel "weather-app2/v2/app/model"
	conf "weather-app2/v2/db/helper"
)

type Root struct {
	LocationalParentGAI string
	FunctionalParentGAI string

	Config *appmodel.Configuration
}

func (r *Root) GetName() string {
	return "weather_app2"
}

func (r *Root) GetDescription() string {
	return "Root asset for Weather App"
}

func (r *Root) GetAssetType() string {
	return "weather_app_root"
}

func (r *Root) GetGAI() string {
	return r.GetAssetType()
}

func (r *Root) GetAssetID() (*int32, error) {
	return conf.GetRootAssetId(context.Background(), r.GetGAI(), *r.Config)
}

func (r *Root) SetAssetID(assetID int32) error {
	rootAsset := appmodel.RootAsset{
		Config:  *r.Config,
		AssetID: assetID,
		Gai:     r.GetGAI(),
	}
	return conf.UpsertRootAsset(context.Background(), rootAsset)
}

func (r *Root) GetLocationalParentGAI() string {
	return r.LocationalParentGAI
}

func (r *Root) GetFunctionalParentGAI() string {
	return r.FunctionalParentGAI
}
func (r *Root) GetSiteID() string {
	return r.Config.SiteId
}

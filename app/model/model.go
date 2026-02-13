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

package appmodel

import "github.com/google/uuid"

type Configuration struct {
	Id              int64     `json:"id" db:"id"`
	TenantId        uuid.UUID `json:"tenant_id" db:"tenant_id"`
	SiteId          string    `json:"site_id" db:"site_id"`
	ApiKey          string    `json:"api_key" db:"api_key"`
	RefreshInterval int32     `json:"refresh_interval" db:"refresh_interval"`
	RequestTimeout  int32     `json:"request_timeout" db:"request_timeout"`
	Enable          bool      `json:"enable" db:"enable"`
	Active          bool      `json:"active" db:"active"`
	UserId          string    `json:"user_id" db:"user_id"`
}

type Asset struct {
	ID              int64   `json:"id" db:"id"`
	ConfigurationId int64   `json:"configuration_id" db:"configuration_id"`
	LocationName    string  `json:"location_name" db:"location_name"`
	Lat             float64 `json:"lat" db:"lat"`
	Lon             float64 `json:"lon" db:"lon"`
	AssetID         int32   `json:"asset_id" db:"asset_id"`
}

type RootAsset struct {
	ID      int64         `json:"id" db:"id"`
	Config  Configuration `json:"config" db:"-"`
	AssetID int32         `json:"asset_id" db:"asset_id"`
	Gai     string        `json:"gai" db:"gai"`
}

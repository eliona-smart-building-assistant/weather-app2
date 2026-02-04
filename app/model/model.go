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
	Id              int64
	TenantId        uuid.UUID
	SiteId          string
	ApiKey          string
	RefreshInterval int32
	RequestTimeout  int32
	Enable          bool
	Active          bool
	UserId          string
}

type Asset struct {
	ID           int64
	ProjectID    string
	LocationName string
	Lat          float64
	Lon          float64
	AssetID      int32
}

type RootAsset struct {
	ID      int64
	Config  Configuration
	AssetID int32
	Gai     string
}

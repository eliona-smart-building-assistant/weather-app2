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

package main

import (
	"time"
	"weather-app2/app"
	dbhelper "weather-app2/db/helper"

	elionaapp "github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

// The main function starts the app by starting all services necessary for this app and waits
// until all services are finished.
func main() {
	log.Info("main", "Starting the app.")

	// Set default database to use boil.*G functions.
	database := db.Database(elionaapp.AppName())
	dbhelper.InitDB(database)
	defer dbhelper.CloseDB()

	// Necessary to close used init resources, because db.Pool() is used in this app.
	defer db.ClosePool()

	// Initialize the app
	app.Initialize()

	// Starting the service to collect the data for this app.
	common.WaitForWithOs(
		common.Loop(app.CollectData, 5*time.Second),
		app.ListenApi,
		app.ListenForOutputChanges,
		common.Loop(app.Heartbeat, 2*time.Minute),
	)

	log.Info("main", "Terminate the app.")
}

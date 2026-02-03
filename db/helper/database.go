//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
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

	"github.com/aarondl/sqlboiler/v4/boil"
	elionabackend "github.com/eliona-smart-building-assistant/backend-frm/pkg/eliona"
	"github.com/eliona-smart-building-assistant/backend-frm/pkg/postgres"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

var (
	sqldb *sql.DB
)

func GetDB() *sql.DB {
	return sqldb
}

// InitDefaultDB Creates the db which can be retrieved with GetDB function
func InitDefaultDB() *postgres.Pool {
	pool := openPool()
	sqldb = pool.StdlibDB()
	boil.SetDB(sqldb)
	return pool
}

func openPool() *postgres.Pool {
	pool, err := elionabackend.GetDatabasePoolWithOverrideRole(context.Background(), "eliona", "weather-app2", 10, postgres.WithResetOnAcquire())
	if err != nil {
		log.Fatal("Database", "Cannot open database: %v", err)
	}
	return pool
}

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
	"database/sql"
	"os"
	"path/filepath"

	"github.com/eliona-smart-building-assistant/go-utils/log"
	_ "modernc.org/sqlite"
)

var (
	sqldb *sql.DB
)

func GetDB() *sql.DB {
	return sqldb
}

// InitDefaultDB Creates the SQLite db which can be retrieved with GetDB function
func InitDefaultDB() *sql.DB {
	db := openDB()
	sqldb = db

	// Initialize the database schema
	if err := InitializeSchema(db); err != nil {
		log.Fatal("Database", "Cannot initialize database schema: %v", err)
	}

	return db
}

func openDB() *sql.DB {
	// Get database path from environment variable, or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		// Default local development path
		dataDir := "data"
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatal("Database", "Cannot create data directory: %v", err)
		}
		dbPath = filepath.Join(dataDir, "weather_app2.db")
	} else {
		// Ensure the directory exists for the database file
		dbDir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatal("Database", "Cannot create database directory: %v", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=1")
	if err != nil {
		log.Fatal("Database", "Cannot open SQLite database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Database", "Cannot connect to SQLite database: %v", err)
	}

	log.Info("Database", "Connected to SQLite database at %s", dbPath)
	return db
}

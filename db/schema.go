//  This file is part of the eliona project.
//  Copyright © 2025 IoTEC AG. All Rights Reserved.

package dbhelper

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

//go:embed schema.sql
var schemaFS embed.FS

// InitializeSchema creates the database tables if they don't exist
func InitializeSchema(db *sql.DB) error {

	// Check if the schema is already initialized
	if isSchemaInitialized(db) {
		log.Info("Database", "Database schema already initialized")
		return nil
	}

	schemaSQL, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("reading schema file: %v", err)
	}

	// Split the SQL file into individual statements
	statements := strings.Split(string(schemaSQL), ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		_, err := db.Exec(statement)
		if err != nil {
			log.Error("Database", "Error executing schema statement: %v\nStatement: %s", err, statement)
			return fmt.Errorf("executing schema statement: %v", err)
		}
	}

	log.Info("Database", "Database schema initialized successfully")
	return nil
}

// isSchemaInitialized checks if the required tables exist
func isSchemaInitialized(db *sql.DB) bool {
	// Check if our main tables exist
	requiredTables := []string{"configuration", "asset", "root_asset"}

	for _, table := range requiredTables {
		var count int
		query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
		err := db.QueryRow(query, table).Scan(&count)
		if err != nil || count == 0 {
			return false
		}
	}

	return true
}

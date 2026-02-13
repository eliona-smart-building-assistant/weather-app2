-- SQLite3 schema for weather-app2
-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Configuration table
CREATE TABLE IF NOT EXISTS configuration (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id TEXT NOT NULL UNIQUE,
    site_id TEXT,
    refresh_interval INTEGER NOT NULL DEFAULT 60,
    request_timeout INTEGER NOT NULL DEFAULT 120,
    active BOOLEAN NOT NULL DEFAULT false,
    enable BOOLEAN NOT NULL DEFAULT false,
    user_id TEXT NOT NULL
);

-- Asset table
CREATE TABLE IF NOT EXISTS asset (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_name TEXT,
    lat REAL,
    lon REAL,
    asset_id INTEGER
);

-- Root Asset table
CREATE TABLE IF NOT EXISTS root_asset (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    configuration_id INTEGER NOT NULL,
    gai TEXT NOT NULL,
    asset_id INTEGER NOT NULL UNIQUE,
    FOREIGN KEY (configuration_id) REFERENCES configuration(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_configuration_tenant_id ON configuration(tenant_id);
CREATE INDEX IF NOT EXISTS idx_asset_location_name ON asset(location_name);
CREATE INDEX IF NOT EXISTS idx_root_asset_configuration_id ON root_asset(configuration_id);
CREATE INDEX IF NOT EXISTS idx_root_asset_asset_id ON root_asset(asset_id);

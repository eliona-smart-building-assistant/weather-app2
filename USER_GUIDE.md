# Weather App User Guide

## Introduction
The Weather App integrates OpenWeatherMap data with Eliona, providing real-time weather information for location-based analytics and energy optimization.

## Overview
This comprehensive guide covers installation, configuration, asset management, and troubleshooting for the Weather App.

## Installation
1. Navigate to the Eliona App Store
2. Search for "Weather App"
3. Click "Install" and confirm the installation
4. Wait for the installation confirmation (typically 1-2 minutes)

## Configuration

### OpenWeatherMap Registration
1. Visit [OpenWeatherMap](https://openweathermap.org/) and create an account
2. Subscribe to the "One Call API 3.0" service (free tier available)
3. Generate your API key in the Account Dashboard
4. Note your API key for Eliona configuration

### App Configuration Parameters

| Parameter | Description | Required | Default Value |
|-----------|-------------|----------|---------------|
| `apiKey` | Your OpenWeatherMap API key | Yes | - |
| `enable` | Enable/disable the configuration | Yes | true |
| `refreshInterval` | Data synchronization interval (seconds) | Yes | 300 |
| `requestTimeout` | API request timeout (seconds) | No | 120 |
| `projectIDs` | Eliona project IDs for data collection | Yes | - |

### Configuration Steps
1. Navigate to `Settings > Apps > Weather` in Eliona
2. Access the Generic Frontend interface
3. Use the config endpoint with PUT method
4. Enter your configuration in JSON format:

```json
{
  "apiKey": "your-api-key-here",
  "enable": true,
  "refreshInterval": 300,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

## Asset Management

### Asset Type
The app creates a `Weather` asset type representing weather monitoring locations.

### Creating Weather Assets
1. Navigate to Assets in Eliona
2. Click "Create New Asset"
3. Select "Weather" as the asset type
4. Configure basic asset properties
5. Save the new asset

### Configuring Location
1. Open your weather asset
2. Click the edit button
3. In the "More Info" section:
   - Enter the location name (e.g., "Zurich, Switzerland")
   - For best results, use format: "City, Country" or "City, State, Country"
4. Save the configuration
5. Refresh the page to verify:
   - Location name appears with additional geographic details
   - Weather data begins populating the asset attributes

### Location Troubleshooting
If the location isn't found:
1. Try more specific location names
2. Include country or state information
3. Verify the location exists in OpenWeatherMap's database
4. Check for typos in the location name

## Weather Data Attributes

The app provides these weather properties:

| Attribute | Description | Unit |
|-----------|-------------|------|
| temperature | Current air temperature | °C |
| feels_like | Perceived temperature accounting for wind and humidity | °C |
| pressure | Atmospheric pressure at the location | hPa |
| humidity | Relative humidity percentage | % |
| dew_point | Temperature at which dew forms | °C |
| uvi | Ultraviolet index indicating sun exposure risk | - |
| clouds | Percentage of cloud cover | % |
| wind_speed | Current wind speed | m/s |
| wind_deg | Wind direction in degrees (0-360) | ° |

## App Status Monitoring

The app creates a "Weather Root" asset that provides:

### Status Indicators
- **Asset Status**: Shows Active/Inactive state
- **Status Attribute**: Current operational status with possible values:

| Status | Description | Recommended Action |
|--------|-------------|-------------------|
| OK | Normal operation | None required |
| API_ERROR | OpenWeatherMap API issue | Verify API key and network connection |
| CONFIG_ERROR | Configuration problem | Review app configuration |
| RATE_LIMIT | API rate limit exceeded | Increase refresh interval or upgrade API plan |
| NETWORK_ERROR | Connection problem | Check network settings |

## Use Cases

### Building Energy Optimization
- Adjust HVAC systems based on outdoor temperature
- Optimize cooling systems using humidity data
- Implement natural ventilation strategies based on wind conditions

### Renewable Energy Management
- Predict solar generation using cloud cover data
- Adjust wind turbine operations based on wind speed/direction
- Implement weather-aware energy storage strategies

### Facility Management
- Plan maintenance activities around weather conditions
- Implement weather-based cleaning schedules
- Optimize landscape irrigation based on precipitation forecasts

### Analytics and Reporting
- Correlate energy usage with weather patterns
- Generate weather impact reports
- Create predictive models for energy consumption

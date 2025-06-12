# Weather User Guide

### Introduction

> The Weather app provides Eliona with weather data from OpenWeatherMap.

## Overview

This guide provides instructions on configuring, installing, and using the Weather app to manage resources and get weather data from OpenWeatherMap.

## Installation

Install the Weather app via the Eliona App Store.

## Configuration

The Weather app requires configuration through Eliona’s settings interface. Below are the general steps and details needed to configure the app.

### Registering in OpenWeatherMap

Create credentials in [OpenWeatherMap](openweathermap.org). Subscribe to One Call API 3.0 (free for most types of usage in Eliona) or other subscription allowing access to current weather data.

Create an API key and save it to use in the Eliona app configuration (section below).

### Configure the Weather app

Configurations can be created in Eliona under `Settings > Apps > Weather` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the config endpoint with the PUT method. Configuration requires the following data:

| Attribute         | Description                                                                     |
|-------------------|---------------------------------------------------------------------------------|
| `apiKey`          | OpenWeatherMap API key obtained in the previous step.                          |
| `enable`          | Flag to enable or disable this configuration.                                   |
| `refreshInterval` | Interval in seconds for data synchronization.                                   |
| `requestTimeout`  | API query timeout in seconds.                                                   |
| `projectIDs`      | List of Eliona project IDs for data collection.                                 |

Example configuration JSON:

```json
{
  "apiKey": "random-cl13nt-s3cr3t",
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

## Asset Creation

Once configured, the app creates a `weather-app-weather` asset type. You can create any number of assets of this asset type, each representing a location to be provided with weather.

## Configuring weather location

With the aforementioned assets, you can specify the location. Go to the asset, click the edit button, and set the location name in "more info" section. After saving, you can refresh the page, and you should see (under "more info" section) the location you input along with state and country information, to confirm that the app found the correct location. If not, please be more specific in the location name and try again.

The asset will then be provided with current weather for the location, which could be used in analytics, energy optimizations and so on.

## App status monitoring

Along with asset creation, an asset called "Weather root" is also created. It's purpose is to inform users of the app status -- It signalizes whether the app is running (Asset status -> Active/Inactive) and it's status - the Status attribute. If the app status is not "OK", it signifies that the app might not be functioning properly. If the error state persists, let us know by submitting a bug report.

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

package broker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	appmodel "weather-app2/app/model"
)

func TestAuthentication(config appmodel.Configuration) error {
	_, err := getGeolocation("Winterthur", config.ApiKey)
	return err
}

func Locate(config appmodel.Configuration, name string) (Geolocation, error) {
	locs, err := getGeolocation(name, config.ApiKey)
	if err != nil {
		return Geolocation{}, fmt.Errorf("getting location: %v", err)
	}
	if len(locs) == 0 {
		return Geolocation{}, fmt.Errorf("location not found")
	}
	return locs[0], err
}

type Geolocation struct {
	Name       string            `json:"name"`
	LocalNames map[string]string `json:"local_names"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state"`
}

type WeatherData struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Timezone string  `json:"timezone"`
	Current  struct {
		Dt   int `json:"dt"`
		Temp struct {
			Day   float64 `json:"day"`
			Min   float64 `json:"min"`
			Max   float64 `json:"max"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"temp"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"current"`
}

func getGeolocation(location string, apiKey string) ([]Geolocation, error) {
	baseURL := "http://api.openweathermap.org/geo/1.0/direct"
	params := url.Values{}
	params.Add("q", location)
	params.Add("limit", "10")
	params.Add("appid", apiKey)

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unsuccessful response: %s: %v", resp.Status, string(body))
	}

	var geolocations []Geolocation
	err = json.Unmarshal(body, &geolocations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return geolocations, nil
}

func getWeather(lat, lon float64, apiKey string) (*WeatherData, error) {
	baseURL := "https://api.openweathermap.org/data/3.0/onecall"
	params := url.Values{}
	params.Add("lat", fmt.Sprintf("%f", lat))
	params.Add("lon", fmt.Sprintf("%f", lon))
	params.Add("exclude", "minutely,hourly,daily,alerts")
	params.Add("appid", apiKey)

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unsuccessful response: %s: %v", resp.Status, string(body))
	}

	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &weatherData, nil
}

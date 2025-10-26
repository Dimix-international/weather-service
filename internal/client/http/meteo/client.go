package meteo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dimix-international/weather-service/internal/config"
)

type client struct {
	httpClient *http.Client
	url        string
	enabled    bool
}

type Weather struct {
	Name string `json:"name"`
	Main struct {
		Kelvin   float64 `json:"temp"`
		Pressure int     `json:"pressure"`
		Humidity float64 `json:"humidity"`
	} `json:"main"`
}

type WeatherStore interface {
	GetWeatherByCoord(lat float64, lon float64) (Weather, error)
	GetWeatherByCity(city string) (Weather, error)
}

func NewClient(config *config.Config, httpClient *http.Client) *client {
	client := &client{
		httpClient: httpClient,
	}

	if config.OpenWeatherKey != "" && config.OpenWeatherURL != "" {
		client.url = fmt.Sprintf("%s?appid=%s", config.OpenWeatherURL, config.OpenWeatherKey)
		client.enabled = true
	}

	return client
}

func (c *client) GetWeatherByCity(city string) (Weather, error) {
	if !c.enabled {
		return Weather{}, nil
	}

	res, err := http.Get(fmt.Sprintf("%s&q=%s", c.url, city))
	if err != nil {
		return Weather{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Weather{}, fmt.Errorf("status code %d", res.StatusCode)
	}

	var weather Weather

	if err = json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return Weather{}, err
	}

	return weather, nil
}

func (c *client) GetWeatherByCoord(lat, lon float64) (Weather, error) {
	if !c.enabled {
		return Weather{}, nil
	}

	res, err := http.Get(fmt.Sprintf("%s&lat=%f&lon=%f", c.url, lat, lon))
	if err != nil {
		return Weather{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Weather{}, fmt.Errorf("status code %d", res.StatusCode)
	}

	var weather Weather

	if err = json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return Weather{}, err
	}

	return weather, nil
}

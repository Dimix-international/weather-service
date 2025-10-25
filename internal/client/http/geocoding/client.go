package geocoding

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dimix-international/weather-service/internal/config"
)

type client struct {
	httpClient *http.Client
	url        string
}

type Response struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewClient(config *config.Config, httpClient *http.Client) *client {
	return &client{
		url:        config.GeoURL,
		httpClient: httpClient,
	}
}

func (c *client) GetCoords(city string) (Response, error) {
	res, err := c.httpClient.Get(
		fmt.Sprintf("%s?name=%s&count=1&language=ru&format=json", c.url, city),
	)

	if err != nil {
		return Response{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("status code %d", res.StatusCode)
	}

	var geoResp struct {
		Results []Response `json:"results"`
	}

	if err := json.NewDecoder(res.Body).Decode(&geoResp); err != nil {
		return Response{}, err
	}

	if len(geoResp.Results) == 0 {
		return Response{}, nil
	}
	return geoResp.Results[0], nil
}

package open_meteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Response struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
	}
}
type client struct {
	httpClient http.Client
}

func NewClient(httpClient http.Client) *client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetTemperature(ctx context.Context, lat, long float64) (Response, error) {
	endpoint := "https://api.open-meteo.com/v1/forecast"
	q := url.Values{}
	q.Set("latitude", fmt.Sprintf("%f", lat))
	q.Set("longitude", fmt.Sprintf("%f", long))
	q.Set("current", "temperature_2m")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "weather-service/1.0")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("Status code: %d", res.StatusCode)
	}

	var response Response
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return Response{}, err
	}

	return response, nil
}

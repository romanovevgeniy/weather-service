package geocoding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Response struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type client struct {
	httpClient http.Client
}

func NewClient(httpClient http.Client) *client {
	return &client{
		httpClient: httpClient,
	}
}

func (c client) GetCoords(ctx context.Context, city string) (Response, error) {
	if city == "" {
		return Response{}, errors.New("city is required")
	}

	endpoint := "https://geocoding-api.open-meteo.com/v1/search"
	q := url.Values{}
	q.Set("name", city)
	q.Set("count", "1")
	q.Set("language", "ru")
	q.Set("format", "json")

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
		return Response{}, fmt.Errorf("Status code: %s", res.Status)
	}

	var geoResp struct {
		Results []Response `json:"results"`
	}

	err = json.NewDecoder(res.Body).Decode(&geoResp)
	if err != nil {
		return Response{}, err
	}

	if len(geoResp.Results) == 0 {
		return Response{}, errors.New("no results")
	}
	return geoResp.Results[0], nil
}

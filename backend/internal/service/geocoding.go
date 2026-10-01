package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GeocodeResult is the structured result of geocoding a query, used to fill a
// salon's location columns (city, postal code, country and coordinates).
type GeocodeResult struct {
	Latitude  float64
	Longitude float64
	City      string
	Postcode  string
	Country   string
}

// nominatimResult is a single result from the Nominatim search API.
type nominatimResult struct {
	Lat     string           `json:"lat"`
	Lon     string           `json:"lon"`
	Address nominatimAddress `json:"address"`
}

type nominatimAddress struct {
	City         string `json:"city"`
	Town         string `json:"town"`
	Village      string `json:"village"`
	Hamlet       string `json:"hamlet"`
	Municipality string `json:"municipality"`
	Postcode     string `json:"postcode"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
}

// GeocodeNominatim resolves a free-text query to coordinates and structured
// address components using OpenStreetMap's Nominatim (free, keyless). It is
// best-effort: on any error it returns ok=false and no error so callers can
// persist without geo.
func GeocodeNominatim(ctx context.Context, query string) (GeocodeResult, bool) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "jsonv2")
	q.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://nominatim.openstreetmap.org/search?"+q.Encode(), nil)
	if err != nil {
		return GeocodeResult{}, false
	}
	// Nominatim's usage policy requires a descriptive User-Agent.
	req.Header.Set("User-Agent", "fejd-app/1.0 (salon booking; geocoding)")
	req.Header.Set("Accept-Language", "sr,en")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GeocodeResult{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GeocodeResult{}, false
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil || len(results) == 0 {
		return GeocodeResult{}, false
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return GeocodeResult{}, false
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return GeocodeResult{}, false
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return GeocodeResult{}, false
	}

	addr := results[0].Address
	country := strings.ToUpper(strings.TrimSpace(addr.CountryCode))
	if country == "" {
		country = strings.TrimSpace(addr.Country)
	}

	return GeocodeResult{
		Latitude:  lat,
		Longitude: lon,
		City:      firstNonEmpty(addr.City, addr.Town, addr.Village, addr.Hamlet, addr.Municipality),
		Postcode:  strings.TrimSpace(addr.Postcode),
		Country:   country,
	}, true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

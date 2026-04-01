// Package geocoder provides reverse geocoding of GPS coordinates using
// Google Maps Geocoding API or OpenStreetMap Nominatim as a free fallback.
package geocoder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ReverseGeocodeResult holds the parsed result of a reverse geocoding lookup.
type ReverseGeocodeResult struct {
	FormattedAddress string
	LocationName     string
	Street           string
	City             string
	State            string
	ZipCode          string
	PostalCode       string
	Country          string
	GooglePlaceId    string
	GoogleMapsUrl    string
}

// Geocoder reverse-geocodes GPS coordinates to human-readable addresses.
type Geocoder struct {
	apiKey    string
	client    *http.Client
	mu        sync.Mutex
	lastCall  time.Time
	rateLimit time.Duration
}

// NewGeocoder creates a new Geocoder. If apiKey is non-empty, Google Maps
// Geocoding API is used; otherwise OpenStreetMap Nominatim is used with a
// 1-request-per-second rate limit (per their usage policy).
func NewGeocoder(apiKey string) *Geocoder {
	rateLimit := time.Duration(0)
	if apiKey == "" {
		rateLimit = time.Second // Nominatim policy: max 1 req/s
	}
	return &Geocoder{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		rateLimit: rateLimit,
	}
}

// ReverseGeocode converts latitude/longitude to a street address.
// Uses Google Maps if an API key was provided, otherwise falls back to
// OpenStreetMap Nominatim.
func (g *Geocoder) ReverseGeocode(lat, lng float64) (*ReverseGeocodeResult, error) {
	g.throttle()

	if g.apiKey != "" {
		return g.reverseGeocodeGoogle(lat, lng)
	}
	return g.reverseGeocodeNominatim(lat, lng)
}

// throttle enforces the rate limit between requests.
func (g *Geocoder) throttle() {
	if g.rateLimit == 0 {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	elapsed := time.Since(g.lastCall)
	if elapsed < g.rateLimit {
		time.Sleep(g.rateLimit - elapsed)
	}
	g.lastCall = time.Now()
}

// reverseGeocodeGoogle calls the Google Maps Geocoding API.
func (g *Geocoder) reverseGeocodeGoogle(lat, lng float64) (*ReverseGeocodeResult, error) {
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f&key=%s",
		lat, lng, g.apiKey,
	)

	resp, err := g.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("google geocode request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading google response: %w", err)
	}

	var gResp googleGeocodeResponse
	if err := json.Unmarshal(body, &gResp); err != nil {
		return nil, fmt.Errorf("parsing google response: %w", err)
	}

	if gResp.Status != "OK" || len(gResp.Results) == 0 {
		return nil, fmt.Errorf("google geocode returned status: %s", gResp.Status)
	}

	first := gResp.Results[0]
	result := &ReverseGeocodeResult{
		FormattedAddress: first.FormattedAddress,
	}

	// Extract place_id and construct Google Maps URL.
	if first.PlaceID != "" {
		result.GooglePlaceId = first.PlaceID
		result.GoogleMapsUrl = fmt.Sprintf("https://www.google.com/maps/place/?q=place_id:%s", first.PlaceID)
	}

	// Extract address components.
	var streetNumber, streetRoute string
	for _, comp := range first.AddressComponents {
		for _, t := range comp.Types {
			switch t {
			case "country":
				result.Country = comp.LongName
			case "locality":
				result.City = comp.LongName
			case "administrative_area_level_1":
				result.State = comp.LongName
			case "postal_code":
				result.ZipCode = comp.LongName
				result.PostalCode = comp.LongName
			case "street_number":
				streetNumber = comp.LongName
			case "route":
				streetRoute = comp.LongName
			case "point_of_interest", "establishment":
				if result.LocationName == "" {
					result.LocationName = comp.LongName
				}
			}
		}
	}

	// Combine street number and route into full street address.
	switch {
	case streetNumber != "" && streetRoute != "":
		result.Street = streetNumber + " " + streetRoute
	case streetRoute != "":
		result.Street = streetRoute
	case streetNumber != "":
		result.Street = streetNumber
	}

	// If no specific location name, use the city.
	if result.LocationName == "" {
		result.LocationName = result.City
	}

	return result, nil
}

// reverseGeocodeNominatim calls the OpenStreetMap Nominatim API (free, rate-limited).
func (g *Geocoder) reverseGeocodeNominatim(lat, lng float64) (*ReverseGeocodeResult, error) {
	url := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f",
		lat, lng,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating nominatim request: %w", err)
	}
	req.Header.Set("User-Agent", "UteamUP-CLI/1.0 (image-analyzer)")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading nominatim response: %w", err)
	}

	var nResp nominatimResponse
	if err := json.Unmarshal(body, &nResp); err != nil {
		return nil, fmt.Errorf("parsing nominatim response: %w", err)
	}

	if nResp.Error != "" {
		return nil, fmt.Errorf("nominatim error: %s", nResp.Error)
	}

	// Construct street from house number + road.
	street := nResp.Address.Road
	if nResp.Address.HouseNumber != "" && street != "" {
		street = nResp.Address.HouseNumber + " " + street
	} else if nResp.Address.HouseNumber != "" {
		street = nResp.Address.HouseNumber
	}

	// Use city, falling back to town or village.
	city := nResp.Address.City
	if city == "" {
		city = nResp.Address.Town
	}
	if city == "" {
		city = nResp.Address.Village
	}

	result := &ReverseGeocodeResult{
		FormattedAddress: nResp.DisplayName,
		Country:          nResp.Address.Country,
		City:             city,
		Street:           street,
		State:            nResp.Address.State,
		ZipCode:          nResp.Address.Postcode,
		PostalCode:       nResp.Address.Postcode,
	}

	// Derive location name from available fields.
	switch {
	case nResp.Address.Building != "":
		result.LocationName = nResp.Address.Building
	case nResp.Address.Amenity != "":
		result.LocationName = nResp.Address.Amenity
	case nResp.Name != "":
		result.LocationName = nResp.Name
	case nResp.Address.City != "":
		result.LocationName = nResp.Address.City
	case nResp.Address.Town != "":
		result.LocationName = nResp.Address.Town
	case nResp.Address.Village != "":
		result.LocationName = nResp.Address.Village
	}

	return result, nil
}

// --- Google API types ---

type googleGeocodeResponse struct {
	Status  string                `json:"status"`
	Results []googleGeocodeResult `json:"results"`
}

type googleGeocodeResult struct {
	FormattedAddress  string                   `json:"formatted_address"`
	PlaceID           string                   `json:"place_id"`
	AddressComponents []googleAddressComponent `json:"address_components"`
}

type googleAddressComponent struct {
	LongName string   `json:"long_name"`
	Types    []string `json:"types"`
}

// --- Nominatim API types ---

type nominatimResponse struct {
	DisplayName string           `json:"display_name"`
	Name        string           `json:"name"`
	Error       string           `json:"error"`
	Address     nominatimAddress `json:"address"`
}

type nominatimAddress struct {
	Building    string `json:"building"`
	Amenity     string `json:"amenity"`
	Road        string `json:"road"`
	HouseNumber string `json:"house_number"`
	City        string `json:"city"`
	Town        string `json:"town"`
	Village     string `json:"village"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
}

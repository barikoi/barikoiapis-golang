package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Point and coordinates string formats shared by several endpoints. These
// mirror the regexes of the TypeScript SDK (src/schema.ts).
var (
	// rePoint matches "latitude,longitude", e.g. "23.8103,90.4125".
	rePoint = regexp.MustCompile(`^-?\d+\.?\d*,-?\d+\.?\d*$`)
	// reCoordinates matches a semicolon-separated list of at least two
	// "longitude,latitude" pairs, e.g. "90.4125,23.8103;90.3742,23.7461".
	reCoordinates = regexp.MustCompile(`^-?\d+\.?\d*,-?\d+\.?\d*(;-?\d+\.?\d*,-?\d+\.?\d*)+$`)
)

// validateCoords checks latitude/longitude bounds before any HTTP call.
func validateCoords(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return ErrInvalidLongitude
	}
	return nil
}

// requireString returns a *ValidationError if value is blank.
func requireString(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

// validatePointString checks a "latitude,longitude" string for presence,
// format, and coordinate bounds. field names the request field in errors.
func validatePointString(field, v string) error {
	if err := requireString(field, v); err != nil {
		return err
	}
	if !rePoint.MatchString(v) {
		return &ValidationError{Field: field, Message: `must be in format "latitude,longitude"`}
	}
	parts := strings.SplitN(v, ",", 2)
	lat, _ := strconv.ParseFloat(parts[0], 64)
	lon, _ := strconv.ParseFloat(parts[1], 64)
	return validateCoords(lat, lon)
}

// validateCoordinatesString checks a "lon,lat;lon,lat" route string for
// presence, format, and per-pair coordinate bounds.
func validateCoordinatesString(v string) error {
	if err := requireString("coordinates", v); err != nil {
		return err
	}
	if !reCoordinates.MatchString(v) {
		return &ValidationError{Field: "coordinates", Message: `must be in format "lon,lat;lon,lat"`}
	}
	for _, pair := range strings.Split(v, ";") {
		parts := strings.SplitN(pair, ",", 2)
		lon, _ := strconv.ParseFloat(parts[0], 64)
		lat, _ := strconv.ParseFloat(parts[1], 64)
		if err := validateCoords(lat, lon); err != nil {
			return err
		}
	}
	return nil
}

// validateEnum checks field against the allowed values, mirroring the Zod
// enums validated client-side by the TypeScript SDK.
func validateEnum(field, v string, allowed ...string) error {
	for _, a := range allowed {
		if v == a {
			return nil
		}
	}
	return &ValidationError{Field: field, Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", "))}
}

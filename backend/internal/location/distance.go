// Package location provides the geographic distance helper used by the
// channel module for proximity discovery. User/channel location tracking
// was removed after the PostgreSQL refactor (2026-10): the app collects GPS
// coordinates client-side and sends them in the channel API payloads.
package location

import "math"

// EarthRadiusKm is the mean Earth radius used by the Haversine formula.
const EarthRadiusKm = 6371.0

// CalculateDistance returns the great-circle distance in kilometers between
// two coordinates using the Haversine formula.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

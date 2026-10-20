package util

import "math"

// AzimuthDegrees returns the initial compass bearing in degrees (0-360, clockwise
// from true north) to travel from (lat1, lon1) to (lat2, lon2) along the
// great-circle path.
func AzimuthDegrees(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	y := math.Sin(deltaLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLon)

	theta := math.Atan2(y, x) * 180 / math.Pi

	return math.Mod(theta+360, 360)
}

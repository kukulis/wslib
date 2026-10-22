package util

import (
	"math"
	"testing"
)

func TestHaversineDistanceKm_SamePoint(t *testing.T) {
	got := HaversineDistanceKm(54.6872, 25.2797, 54.6872, 25.2797)
	if got != 0 {
		t.Errorf("distance to the same point = %v, want 0", got)
	}
}

// One degree of latitude is, by definition, earthRadiusKm*(pi/180) of great-circle
// distance, exactly, regardless of where on Earth it's measured.
func TestHaversineDistanceKm_OneDegreeLatitude(t *testing.T) {
	want := earthRadiusKm * math.Pi / 180 // ~111.19km
	got := HaversineDistanceKm(0, 0, 1, 0)

	if math.Abs(got-want) > 0.1 {
		t.Errorf("distance for 1 degree of latitude = %v km, want %v km", got, want)
	}
}

// At the equator, a degree of longitude covers the same distance as a degree of
// latitude (no meridian convergence yet).
func TestHaversineDistanceKm_OneDegreeLongitudeAtEquator(t *testing.T) {
	want := earthRadiusKm * math.Pi / 180 // ~111.19km
	got := HaversineDistanceKm(0, 0, 0, 1)

	if math.Abs(got-want) > 0.1 {
		t.Errorf("distance for 1 degree of longitude at the equator = %v km, want %v km", got, want)
	}
}

// At 60 degrees latitude, a degree of longitude covers roughly cos(60)=0.5 of what
// it covers at the equator — this is the meridian-convergence correction that
// distinguishes haversine from plain Pythagorean distance on lat/lon.
func TestHaversineDistanceKm_OneDegreeLongitudeAt60North(t *testing.T) {
	want := earthRadiusKm * math.Pi / 180 * math.Cos(60*math.Pi/180) // ~55.6km
	got := HaversineDistanceKm(60, 0, 60, 1)

	if math.Abs(got-want) > 1.0 {
		t.Errorf("distance for 1 degree of longitude at 60N = %v km, want ~%v km", got, want)
	}
}

func TestHaversineDistanceKm_Symmetric(t *testing.T) {
	const vilniusLat, vilniusLon = 54.6872, 25.2797
	const kaunasLat, kaunasLon = 54.8985, 23.9036

	forward := HaversineDistanceKm(vilniusLat, vilniusLon, kaunasLat, kaunasLon)
	backward := HaversineDistanceKm(kaunasLat, kaunasLon, vilniusLat, vilniusLon)

	if forward != backward {
		t.Errorf("distance not symmetric: %v vs %v", forward, backward)
	}
	if forward <= 0 {
		t.Errorf("Vilnius-Kaunas distance = %v km, want > 0", forward)
	}
}

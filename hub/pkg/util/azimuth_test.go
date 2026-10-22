package util

import (
	"math"
	"testing"
)

func TestAzimuthDegrees_DueNorth(t *testing.T) {
	got := AzimuthDegrees(0, 0, 1, 0)
	if math.Abs(got-0) > 1e-9 {
		t.Errorf("azimuth due north = %v, want 0", got)
	}
}

func TestAzimuthDegrees_DueEastAtEquator(t *testing.T) {
	got := AzimuthDegrees(0, 0, 0, 1)
	if math.Abs(got-90) > 1e-9 {
		t.Errorf("azimuth due east at equator = %v, want 90", got)
	}
}

func TestAzimuthDegrees_DueSouth(t *testing.T) {
	got := AzimuthDegrees(1, 0, 0, 0)
	if math.Abs(got-180) > 1e-9 {
		t.Errorf("azimuth due south = %v, want 180", got)
	}
}

func TestAzimuthDegrees_DueWestAtEquator(t *testing.T) {
	got := AzimuthDegrees(0, 1, 0, 0)
	if math.Abs(got-270) > 1e-9 {
		t.Errorf("azimuth due west at equator = %v, want 270", got)
	}
}

// Expected values cross-checked with an independent Python implementation of
// the same formula, since the forward and return bearings between two points
// are not simply 180 degrees apart off the equator.
func TestAzimuthDegrees_VilniusToKaunas(t *testing.T) {
	const vilniusLat, vilniusLon = 54.6872, 25.2797
	const kaunasLat, kaunasLon = 54.8985, 23.9036

	forward := AzimuthDegrees(vilniusLat, vilniusLon, kaunasLat, kaunasLon)
	backward := AzimuthDegrees(kaunasLat, kaunasLon, vilniusLat, vilniusLon)

	if math.Abs(forward-285.4750509439034) > 1e-9 {
		t.Errorf("Vilnius->Kaunas azimuth = %v, want ~285.475", forward)
	}
	if math.Abs(backward-104.35065696915683) > 1e-9 {
		t.Errorf("Kaunas->Vilnius azimuth = %v, want ~104.351", backward)
	}
}

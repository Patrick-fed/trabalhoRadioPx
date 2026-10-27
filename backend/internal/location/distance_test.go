package location

import (
	"testing"
)

func TestCalculateDistance_SamePoint(t *testing.T) {
	distance := CalculateDistance(0, 0, 0, 0)
	if distance != 0 {
		t.Errorf("Expected 0, got %f", distance)
	}
}

func TestCalculateDistance_KnownDistance(t *testing.T) {
	// Distance between São Paulo and Rio de Janeiro is approximately 360-430km
	spLat := -23.5505
	spLng := -46.6333
	rjLat := -22.9068
	rjLng := -43.1729

	distance := CalculateDistance(spLat, spLng, rjLat, rjLng)

	if distance < 300 || distance > 500 {
		t.Errorf("Expected distance between 300-500km, got %f", distance)
	}
}

func TestCalculateDistance_AntipodalPoints(t *testing.T) {
	// Distance from North Pole to South Pole is approximately 20000km
	distance := CalculateDistance(90, 0, -90, 0)

	if distance < 19000 || distance > 21000 {
		t.Errorf("Expected distance around 20000km, got %f", distance)
	}
}

func TestCalculateDistance_Equator(t *testing.T) {
	// Distance along equator from 0 to 1 degree is approximately 111km
	distance := CalculateDistance(0, 0, 0, 1)

	if distance < 100 || distance > 120 {
		t.Errorf("Expected distance around 111km, got %f", distance)
	}
}

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

func TestLocationService_UpdateUserLocation(t *testing.T) {
	service := NewLocationService()

	update := &LocationUpdate{
		Latitude:  -23.5505,
		Longitude: -46.6333,
		Accuracy:  10.0,
	}

	location, err := service.UpdateUserLocation("user1", update)
	if err != nil {
		t.Fatalf("UpdateUserLocation failed: %v", err)
	}

	if location.Latitude != update.Latitude {
		t.Errorf("Expected latitude %f, got %f", update.Latitude, location.Latitude)
	}
	if location.Longitude != update.Longitude {
		t.Errorf("Expected longitude %f, got %f", update.Longitude, location.Longitude)
	}
}

func TestLocationService_GetUserLocation(t *testing.T) {
	service := NewLocationService()

	update := &LocationUpdate{
		Latitude:  -23.5505,
		Longitude: -46.6333,
	}

	service.UpdateUserLocation("user1", update)

	location, exists := service.GetUserLocation("user1")
	if !exists {
		t.Fatal("Expected location to exist")
	}

	if location.Latitude != update.Latitude {
		t.Errorf("Expected latitude %f, got %f", update.Latitude, location.Latitude)
	}
}

func TestLocationService_GetUsersInRadius(t *testing.T) {
	service := NewLocationService()

	service.UpdateUserLocation("user1", &LocationUpdate{
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	service.UpdateUserLocation("user2", &LocationUpdate{
		Latitude:  -23.5605,
		Longitude: -46.6433,
	})

	service.UpdateUserLocation("user3", &LocationUpdate{
		Latitude:  -22.9068,
		Longitude: -43.1729,
	})

	users := service.GetUsersInRadius(-23.5505, -46.6333, 10.0)

	if len(users) != 2 {
		t.Errorf("Expected 2 users in radius, got %d", len(users))
	}
}

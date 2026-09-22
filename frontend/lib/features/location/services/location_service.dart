import 'dart:async';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:geolocator/geolocator.dart';

import '../models/location_model.dart';

class LocationService {
  StreamSubscription<Position>? _positionSubscription;
  final StreamController<LocationModel> _locationController =
      StreamController<LocationModel>.broadcast();

  bool _isTracking = false;
  LocationSettings? _locationSettings;

  // Mock location for desktop (São Paulo)
  static const double _defaultLatitude = -23.5505;
  static const double _defaultLongitude = -46.6333;

  Stream<LocationModel> get locationStream => _locationController.stream;
  bool get isTracking => _isTracking;

  bool get _isDesktop => kIsWeb;

  Future<bool> checkPermission() async {
    if (_isDesktop) return true;

    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      return false;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        return false;
      }
    }

    if (permission == LocationPermission.deniedForever) {
      return false;
    }

    return true;
  }

  Future<void> startTracking({
    int distanceFilter = 10,
    int intervalSeconds = 30,
  }) async {
    if (_isTracking) return;

    if (_isDesktop) {
      _isTracking = true;
      return;
    }

    final hasPermission = await checkPermission();
    if (!hasPermission) {
      throw Exception('Location permission not granted');
    }

    _locationSettings = LocationSettings(
      accuracy: LocationAccuracy.high,
      distanceFilter: distanceFilter,
      timeLimit: Duration(seconds: intervalSeconds),
    );

    _positionSubscription = Geolocator.getPositionStream(
      locationSettings: _locationSettings,
    ).listen(
      (Position position) {
        final location = LocationModel(
          latitude: position.latitude,
          longitude: position.longitude,
          accuracy: position.accuracy,
          timestamp: position.timestamp,
          updatedAt: DateTime.now(),
        );
        _locationController.add(location);
      },
      onError: (error) {
        _locationController.addError(error);
      },
    );

    _isTracking = true;
  }

  Future<void> stopTracking() async {
    await _positionSubscription?.cancel();
    _positionSubscription = null;
    _isTracking = false;
  }

  Future<LocationModel?> getCurrentLocation() async {
    if (_isDesktop) {
      return LocationModel(
        latitude: _defaultLatitude,
        longitude: _defaultLongitude,
        accuracy: 0,
        timestamp: DateTime.now(),
        updatedAt: DateTime.now(),
      );
    }

    final hasPermission = await checkPermission();
    if (!hasPermission) {
      return null;
    }

    try {
      final position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      );

      return LocationModel(
        latitude: position.latitude,
        longitude: position.longitude,
        accuracy: position.accuracy,
        timestamp: position.timestamp,
        updatedAt: DateTime.now(),
      );
    } catch (e) {
      return null;
    }
  }

  double calculateDistance(
    double lat1, double lon1,
    double lat2, double lon2,
  ) {
    return Geolocator.distanceBetween(lat1, lon1, lat2, lon2) / 1000;
  }

  void dispose() {
    stopTracking();
    _locationController.close();
  }
}

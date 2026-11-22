import 'dart:async';

import 'package:flutter/foundation.dart' show visibleForTesting;

import '../features/location/services/location_service.dart';

/// Periodically reports the device position to `POST /api/v1/location`
/// while the user is logged in (RF04 — location updated every 30s, DEC-020).
///
/// Singleton owned by [MainScreen]: start on login, stop on logout and when
/// the app goes to background (mirrors the channels-screen polling pattern).
class LocationReporter {
  LocationReporter._({LocationService? service, Future<void> Function()? report})
      : _service = service,
        _report = report;

  static final LocationReporter instance = LocationReporter._();

  static const Duration interval = Duration(seconds: 30);

  Timer? _timer;
  bool _inFlight = false;
  final LocationService? _service;
  final Future<void> Function()? _report;
  LocationService? _resolvedService;

  @visibleForTesting
  bool get isRunning => _timer != null;

  @visibleForTesting
  static LocationReporter createForTest({Future<void> Function()? report}) {
    return LocationReporter._(report: report);
  }

  /// Starts the periodic report. Idempotent. Sends an immediate report so a
  /// resumed session does not wait a full interval.
  void start() {
    if (_timer != null) return;
    _timer = Timer.periodic(interval, (_) => _tick());
    _tick();
  }

  /// Stops the periodic report. Safe to call when not running.
  void stop() {
    _timer?.cancel();
    _timer = null;
  }

  Future<void> _tick() async {
    if (_inFlight) return; // skip if the previous report is still running
    _inFlight = true;
    try {
      if (_report != null) {
        await _report();
      } else {
        _resolvedService ??= _service ?? LocationService();
        await _resolvedService!.reportPosition();
      }
    } finally {
      _inFlight = false;
    }
  }
}

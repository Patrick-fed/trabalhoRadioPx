import 'dart:async';

import 'package:flutter/foundation.dart';

import '../models/bluetooth_proximity_state.dart';
import 'bluetooth_state_provider.dart';

class BluetoothProximityController extends ChangeNotifier {
  BluetoothProximityController(this._provider) {
    _initialize();
  }

  final BluetoothStateProvider _provider;
  StreamSubscription<BluetoothProximityState>? _subscription;
  BluetoothProximityState _state = BluetoothProximityState.unavailable;
  bool _disposed = false;

  BluetoothProximityState get state => _state;

  Future<void> _initialize() async {
    BluetoothProximityState initialState;
    try {
      initialState = await _provider.readState();
    } catch (_) {
      initialState = BluetoothProximityState.unavailable;
    }
    if (_disposed) {
      return;
    }
    _setState(initialState);

    try {
      _subscription = _provider.watchState().listen(
        _setState,
        onError: (Object error, StackTrace stackTrace) {
          _setState(BluetoothProximityState.unavailable);
        },
      );
    } catch (_) {
      _setState(BluetoothProximityState.unavailable);
    }
  }

  void _setState(BluetoothProximityState state) {
    if (_disposed || _state == state) {
      return;
    }
    _state = state;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    _subscription?.cancel();
    super.dispose();
  }
}

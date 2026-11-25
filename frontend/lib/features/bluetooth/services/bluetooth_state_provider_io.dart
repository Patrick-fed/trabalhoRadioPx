import 'dart:io';

import 'package:bluetooth_manager/bluetooth_manager.dart' as bluetooth_manager;
import 'package:bluetooth_manager/models/bluetooth_models.dart'
    as bluetooth_models;

import '../models/bluetooth_proximity_state.dart';
import 'bluetooth_state_provider.dart';

class PluginBluetoothStateProvider implements BluetoothStateProvider {
  PluginBluetoothStateProvider({bluetooth_manager.BluetoothManager? manager})
      : _manager = manager ?? bluetooth_manager.BluetoothManager();

  final bluetooth_manager.BluetoothManager _manager;

  @override
  Future<BluetoothProximityState> readState() async {
    try {
      return _mapState(await _manager.getBluetoothState());
    } catch (_) {
      return BluetoothProximityState.unavailable;
    }
  }

  @override
  Stream<BluetoothProximityState> watchState() {
    try {
      return _manager.getBluetoothStateStream().map(_mapState).handleError(
            (Object error, StackTrace stackTrace) =>
                BluetoothProximityState.unavailable,
            test: (error) => true,
          );
    } catch (_) {
      return Stream<BluetoothProximityState>.value(
        BluetoothProximityState.unavailable,
      );
    }
  }

  BluetoothProximityState _mapState(bluetooth_models.BluetoothState state) {
    switch (state) {
      case bluetooth_models.BluetoothState.on:
        return BluetoothProximityState.nearby;
      case bluetooth_models.BluetoothState.off:
        return BluetoothProximityState.away;
      case bluetooth_models.BluetoothState.uknow:
        return BluetoothProximityState.unavailable;
    }
  }
}

BluetoothStateProvider createBluetoothStateProvider() {
  if (!Platform.isAndroid) {
    return const UnsupportedBluetoothStateProvider();
  }
  return PluginBluetoothStateProvider();
}

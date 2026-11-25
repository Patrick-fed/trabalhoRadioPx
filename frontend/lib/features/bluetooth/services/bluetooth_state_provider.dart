import '../models/bluetooth_proximity_state.dart';

abstract class BluetoothStateProvider {
  Future<BluetoothProximityState> readState();

  Stream<BluetoothProximityState> watchState();
}

class UnsupportedBluetoothStateProvider implements BluetoothStateProvider {
  const UnsupportedBluetoothStateProvider();

  @override
  Future<BluetoothProximityState> readState() async {
    return BluetoothProximityState.unavailable;
  }

  @override
  Stream<BluetoothProximityState> watchState() {
    return const Stream<BluetoothProximityState>.empty();
  }
}

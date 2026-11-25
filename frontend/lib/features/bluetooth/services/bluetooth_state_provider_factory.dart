import 'bluetooth_state_provider_stub.dart'
    if (dart.library.io) 'bluetooth_state_provider_io.dart'
    as impl;

import 'bluetooth_state_provider.dart';

BluetoothStateProvider createBluetoothStateProvider() {
  return impl.createBluetoothStateProvider();
}

enum BluetoothProximityState {
  nearby('Alguém por perto'),
  away('Ninguém por perto'),
  unavailable('Indisponível');

  const BluetoothProximityState(this.label);

  final String label;
}

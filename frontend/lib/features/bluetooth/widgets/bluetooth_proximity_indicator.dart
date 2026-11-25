import 'package:flutter/material.dart';

import '../models/bluetooth_proximity_state.dart';
import '../services/bluetooth_proximity_controller.dart';

class BluetoothProximityIndicator extends StatelessWidget {
  const BluetoothProximityIndicator({super.key, required this.controller});

  static const Key indicatorKey = Key('bluetooth-proximity-indicator');

  final BluetoothProximityController controller;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: controller,
      builder: (context, _) {
        final state = controller.state;
        final color = _colorFor(state);
        final icon = _iconFor(state);

        return Container(
          width: double.infinity,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            children: [
              Icon(icon, size: 18, color: color),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  state.label,
                  style: TextStyle(color: color, fontSize: 13),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Color _colorFor(BluetoothProximityState state) {
    switch (state) {
      case BluetoothProximityState.nearby:
        return Colors.green;
      case BluetoothProximityState.away:
        return Colors.orange;
      case BluetoothProximityState.unavailable:
        return Colors.grey;
    }
  }

  IconData _iconFor(BluetoothProximityState state) {
    switch (state) {
      case BluetoothProximityState.nearby:
        return Icons.bluetooth_connected;
      case BluetoothProximityState.away:
        return Icons.bluetooth_disabled;
      case BluetoothProximityState.unavailable:
        return Icons.bluetooth;
    }
  }
}

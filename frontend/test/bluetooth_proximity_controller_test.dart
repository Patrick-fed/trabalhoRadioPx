import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/bluetooth/models/bluetooth_proximity_state.dart';
import 'package:radiopx/features/bluetooth/services/bluetooth_proximity_controller.dart';
import 'package:radiopx/features/bluetooth/services/bluetooth_state_provider.dart';

class FakeBluetoothStateProvider implements BluetoothStateProvider {
  FakeBluetoothStateProvider({
    this.initialState = BluetoothProximityState.unavailable,
    this.initialError = false,
  });

  final BluetoothProximityState initialState;
  final bool initialError;
  final StreamController<BluetoothProximityState> _streamController =
      StreamController<BluetoothProximityState>.broadcast();

  bool cancelled = false;

  @override
  Future<BluetoothProximityState> readState() async {
    if (initialError) {
      throw StateError('falha simulada');
    }
    return initialState;
  }

  @override
  Stream<BluetoothProximityState> watchState() {
    _streamController.onCancel = () => cancelled = true;
    return _streamController.stream;
  }

  void emit(BluetoothProximityState state) {
    _streamController.add(state);
  }

  void emitError() {
    _streamController.addError(StateError('falha simulada'));
  }
}

void main() {
  group('BluetoothProximityController (RF10)', () {
    test('usa o estado inicial consultado', () async {
      final provider = FakeBluetoothStateProvider(
        initialState: BluetoothProximityState.nearby,
      );
      final controller = BluetoothProximityController(provider);

      await Future<void>.delayed(Duration.zero);

      expect(controller.state, BluetoothProximityState.nearby);
      controller.dispose();
    });

    test('atualiza o estado quando o stream emite mudanças', () async {
      final provider = FakeBluetoothStateProvider(
        initialState: BluetoothProximityState.nearby,
      );
      final controller = BluetoothProximityController(provider);
      await Future<void>.delayed(Duration.zero);

      provider.emit(BluetoothProximityState.away);
      await Future<void>.delayed(Duration.zero);
      expect(controller.state, BluetoothProximityState.away);

      provider.emit(BluetoothProximityState.nearby);
      await Future<void>.delayed(Duration.zero);
      expect(controller.state, BluetoothProximityState.nearby);

      controller.dispose();
    });

    test('degrada para indisponível quando a leitura falha', () async {
      final provider = FakeBluetoothStateProvider(initialError: true);
      final controller = BluetoothProximityController(provider);

      await Future<void>.delayed(Duration.zero);

      expect(controller.state, BluetoothProximityState.unavailable);
      controller.dispose();
    });

    test('degrada para indisponível quando o stream falha', () async {
      final provider = FakeBluetoothStateProvider(
        initialState: BluetoothProximityState.nearby,
      );
      final controller = BluetoothProximityController(provider);
      await Future<void>.delayed(Duration.zero);

      provider.emitError();
      await Future<void>.delayed(Duration.zero);

      expect(controller.state, BluetoothProximityState.unavailable);
      controller.dispose();
    });

    test('cancela a assinatura ao ser descartado', () async {
      final provider = FakeBluetoothStateProvider();
      final controller = BluetoothProximityController(provider);
      await Future<void>.delayed(Duration.zero);

      controller.dispose();

      expect(provider.cancelled, isTrue);
    });
  });
}

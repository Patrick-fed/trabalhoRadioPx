import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/bluetooth/models/bluetooth_proximity_state.dart';
import 'package:radiopx/features/bluetooth/services/bluetooth_proximity_controller.dart';
import 'package:radiopx/features/bluetooth/services/bluetooth_state_provider.dart';
import 'package:radiopx/features/bluetooth/widgets/bluetooth_proximity_indicator.dart';

class _FakeProvider implements BluetoothStateProvider {
  _FakeProvider(this.state);

  final BluetoothProximityState state;

  @override
  Future<BluetoothProximityState> readState() async => state;

  @override
  Stream<BluetoothProximityState> watchState() =>
      const Stream<BluetoothProximityState>.empty();
}

void main() {
  group('BluetoothProximityIndicator (RF10)', () {
    testWidgets('exibe os três textos esperados', (tester) async {
      for (final entry in BluetoothProximityState.values) {
        final controller = BluetoothProximityController(_FakeProvider(entry));
        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: BluetoothProximityIndicator(controller: controller),
            ),
          ),
        );
        await tester.pumpAndSettle();

        expect(find.text(entry.label), findsOneWidget);
        controller.dispose();
      }
    });

    testWidgets('permanece visível ao alternar Canais, Falar e Perfil', (
      tester,
    ) async {
      final controller = BluetoothProximityController(
        _FakeProvider(BluetoothProximityState.nearby),
      );
      var index = 0;
      const destinations = ['Canais', 'Falar', 'Perfil'];

      await tester.pumpWidget(
        MaterialApp(
          home: StatefulBuilder(
            builder: (context, setState) {
              return Scaffold(
                body: Column(
                  children: [
                    SafeArea(
                      bottom: false,
                      child: Padding(
                        padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                        child: BluetoothProximityIndicator(
                          key: BluetoothProximityIndicator.indicatorKey,
                          controller: controller,
                        ),
                      ),
                    ),
                    Expanded(
                      child: Center(child: Text('tela-${destinations[index]}')),
                    ),
                  ],
                ),
                bottomNavigationBar: NavigationBar(
                  selectedIndex: index,
                  onDestinationSelected: (value) {
                    setState(() => index = value);
                  },
                  destinations: [
                    for (final label in destinations)
                      NavigationDestination(
                        icon: const Icon(Icons.circle),
                        label: label,
                      ),
                  ],
                ),
              );
            },
          ),
        ),
      );
      await tester.pumpAndSettle();

      for (var destinationIndex = 0;
          destinationIndex < destinations.length;
          destinationIndex++) {
        final navigationBar = tester.widget<NavigationBar>(
          find.byType(NavigationBar),
        );
        navigationBar.onDestinationSelected!(destinationIndex);
        await tester.pumpAndSettle();

        final label = destinations[destinationIndex];
        expect(
          find.byKey(BluetoothProximityIndicator.indicatorKey),
          findsOneWidget,
        );
        expect(find.text(BluetoothProximityState.nearby.label), findsOneWidget);
        expect(find.text('tela-$label'), findsOneWidget);
      }

      controller.dispose();
    });
  });
}

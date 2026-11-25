import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

void main() {
  group('Bluetooth isolation (RF10)', () {
    test('feature source does not call forbidden Bluetooth APIs', () {
      final directory = Directory('lib/features/bluetooth');
      expect(directory.existsSync(), isTrue);

      final sources = directory
          .listSync(recursive: true)
          .whereType<File>()
          .where((file) => file.path.endsWith('.dart'))
          .map((file) => file.readAsStringSync())
          .join('\n');

      final forbiddenPatterns = <RegExp>[
        RegExp(r'\benableBluetooth\s*\('),
        RegExp(r'\bdisableBluetooth\s*\('),
        RegExp(r'\bstartDiscovery\s*\('),
        RegExp(r'\bcancelDiscovery\s*\('),
        RegExp(r'\bgetBondedDevices\s*\('),
        RegExp(r'\bcreateBond\s*\('),
        RegExp(r'\bcreateRfcommSocket\s*\('),
        RegExp(r'\bconnect\s*\('),
        RegExp(r'\bdisconnect\s*\('),
        RegExp(r'\bstartAdvertising\s*\('),
        RegExp(r'\bstopAdvertising\s*\('),
        RegExp(r'\blet\s*\('),
        RegExp(r'\brequestPermissions\s*\('),
        RegExp(r'\bcheckSelfPermission\s*\('),
        RegExp(r'permission_handler'),
        RegExp(r'SharedPreferences'),
        RegExp(r'web_socket_channel'),
        RegExp(r'WebSocket'),
        RegExp(r'package:http'),
        RegExp(r'package:dio'),
        RegExp(r'HttpClient'),
      ];

      for (final pattern in forbiddenPatterns) {
        expect(
          sources,
          isNot(matches(pattern)),
          reason: 'Padrão proibido encontrado: ${pattern.pattern}',
        );
      }
    });

    test('provider uses only read and watch plugin calls', () {
      final source = File(
        'lib/features/bluetooth/services/bluetooth_state_provider_io.dart',
      ).readAsStringSync();

      final managerCalls =
          RegExp(r'(?:^|[^A-Za-z0-9_])_manager\.([A-Za-z_]\w*)\s*\(')
              .allMatches(source)
              .map((match) => match.group(1))
              .toSet();

      expect(
        managerCalls,
        equals({'getBluetoothState', 'getBluetoothStateStream'}),
      );
      expect(source, contains('Platform.isAndroid'));
    });
  });
}

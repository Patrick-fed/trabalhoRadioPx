import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/auth/services/auth_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

String _jwt({required int exp}) {
  final header = _b64({'alg': 'HS256', 'typ': 'JWT'});
  final payload = _b64({'sub': 'user-1', 'exp': exp});
  return '$header.$payload.signature';
}

String _b64(Map<String, Object> map) =>
    base64Url.encode(Uint8List.fromList(utf8.encode(jsonEncode(map))));

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  final futureExp =
      DateTime.now().add(const Duration(hours: 1)).millisecondsSinceEpoch ~/ 1000;
  final pastExp = DateTime.now()
          .subtract(const Duration(hours: 1))
          .millisecondsSinceEpoch ~/
      1000;

  group('AuthService.isLoggedIn', () {
    test('false when there is no token', () async {
      SharedPreferences.setMockInitialValues({});
      expect(await AuthService().isLoggedIn(), isFalse);
    });

    test('false and clears storage for a malformed token', () async {
      SharedPreferences.setMockInitialValues({'auth_token': 'not-a-jwt'});

      expect(await AuthService().isLoggedIn(), isFalse);

      final prefs = await SharedPreferences.getInstance();
      expect(prefs.getString('auth_token'), isNull);
    });

    test('true for a well-formed non-expired token', () async {
      SharedPreferences.setMockInitialValues({
        'auth_token': _jwt(exp: futureExp),
      });

      expect(await AuthService().isLoggedIn(), isTrue);
    });

    test('false for an expired token without a refresh token', () async {
      SharedPreferences.setMockInitialValues({
        'auth_token': _jwt(exp: pastExp),
      });

      expect(await AuthService().isLoggedIn(), isFalse);
    });
  });

  group('AuthService.getValidToken', () {
    test('null when there is no session', () async {
      SharedPreferences.setMockInitialValues({});
      expect(await AuthService().getValidToken(), isNull);
    });

    test('returns the stored non-expired token', () async {
      final token = _jwt(exp: futureExp);
      SharedPreferences.setMockInitialValues({'auth_token': token});

      expect(await AuthService().getValidToken(), token);
    });
  });
}

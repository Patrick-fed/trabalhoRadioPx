import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/auth/services/token_validity.dart';

String buildJwt({required int exp}) {
  final header = _base64UrlEncode({'alg': 'HS256', 'typ': 'JWT'});
  final payload = _base64UrlEncode({'sub': 'user-1', 'exp': exp});
  return '$header.$payload.signature';
}

String _base64UrlEncode(Map<String, Object> map) {
  return base64Url.encode(
    Uint8List.fromList(utf8.encode(jsonEncode(map))),
  );
}

void main() {
  group('tokenExpiry', () {
    test('extracts exp from payload', () {
      final expiry = jwtExpiry(buildJwt(exp: 2000000000));
      expect(expiry, DateTime.fromMillisecondsSinceEpoch(2000000000 * 1000));
    });

    test('returns null for malformed token', () {
      expect(jwtExpiry('not-a-jwt'), isNull);
      expect(jwtExpiry('a.b'), isNull);
      expect(jwtExpiry('a.b.c.d'), isNull);
    });

    test('returns null when payload has no exp', () {
      final noExp =
          '${_base64UrlEncode({'alg': 'none'})}.${_base64UrlEncode({'sub': 'u'})}.sig';
      expect(jwtExpiry(noExp), isNull);
    });

    test('returns null for invalid base64', () {
      expect(jwtExpiry('x.@invalid%.sig'), isNull);
    });
  });

  group('tokenIsExpired', () {
    final now = DateTime.fromMillisecondsSinceEpoch(2000000000 * 1000);

    test('true when exp is in the past', () {
      final token = buildJwt(exp: (now.millisecondsSinceEpoch ~/ 1000) - 60);
      expect(tokenIsExpired(token, now: now), isTrue);
    });

    test('false when exp is in the future', () {
      final token = buildJwt(exp: (now.millisecondsSinceEpoch ~/ 1000) + 3600);
      expect(tokenIsExpired(token, now: now), isFalse);
    });

    test('true exactly at exp', () {
      final token = buildJwt(exp: now.millisecondsSinceEpoch ~/ 1000);
      expect(tokenIsExpired(token, now: now), isTrue);
    });

    test('false for malformed token (lenient)', () {
      expect(tokenIsExpired('not-a-jwt', now: now), isFalse);
    });
  });
}
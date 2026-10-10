import 'dart:convert';

DateTime? jwtExpiry(String token) {
  final parts = token.split('.');
  if (parts.length != 3) return null;

  final payload = parts[1].replaceAll('-', '+').replaceAll('_', '/');
  final missingPadding = (4 - payload.length % 4) % 4;
  final padded = payload.padRight(payload.length + missingPadding, '=');

  try {
    final decoded = utf8.decode(base64.decode(padded));
    final map = jsonDecode(decoded);
    if (map is Map && map['exp'] is int) {
      return DateTime.fromMillisecondsSinceEpoch((map['exp'] as int) * 1000);
    }
  } catch (_) {
    return null;
  }

  return null;
}

bool tokenIsExpired(String token, {DateTime? now}) {
  final expiry = jwtExpiry(token);
  if (expiry == null) return false;
  return !(now ?? DateTime.now()).isBefore(expiry);
}
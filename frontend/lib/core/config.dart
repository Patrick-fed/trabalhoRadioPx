import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

class AppConfig {
  AppConfig._();

  static const String _kBuildDefault = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );
  static const String _prefsKey = 'server_url';

  static String _value = _kBuildDefault;

  static String get apiBaseUrl => _value;

  static String get wsBaseUrl {
    if (_value.startsWith('https:')) {
      return _value.replaceFirst('https:', 'wss:');
    }
    return _value.replaceFirst('http:', 'ws:');
  }

  /// Host atual sem o esquema, para exibição (ex.: no perfil).
  static String get displayHost {
    return _value.replaceFirst(RegExp(r'^https?://'), '');
  }

  static String get buildDefault => _kBuildDefault;

  /// Carrega a URL salva pelo usuário (se houver) antes do runApp.
  static Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    final saved = prefs.getString(_prefsKey);
    if (saved != null && saved.isNotEmpty) {
      _value = saved;
    }
  }

  /// Valida, persiste e aplica a nova URL. Lança [ArgumentError] se inválida.
  static Future<void> save(String raw) async {
    final normalized = normalizeUrl(raw);
    if (normalized == null) {
      throw ArgumentError('Informe uma URL começando com http:// ou https://');
    }
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefsKey, normalized);
    _value = normalized;
  }

  /// Trim + remoção de `/` final; só aceita `http://`/`https://`.
  /// Retorna `null` quando a entrada é inválida.
  static String? normalizeUrl(String raw) {
    var url = raw.trim();
    while (url.endsWith('/')) {
      url = url.substring(0, url.length - 1);
    }
    if (url.startsWith('http://') || url.startsWith('https://')) {
      return url;
    }
    return null;
  }

  @visibleForTesting
  static void resetForTest() {
    _value = _kBuildDefault;
  }
}

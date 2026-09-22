class AppConfig {
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  static String get wsBaseUrl {
    if (apiBaseUrl.startsWith('https:')) {
      return apiBaseUrl.replaceFirst('https:', 'wss:');
    }
    return apiBaseUrl.replaceFirst('http:', 'ws:');
  }
}
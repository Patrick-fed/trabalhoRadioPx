class LocationPermissionHelper {
  static Future<bool> checkPermission() async {
    return true;
  }

  static Future<bool> requestPermission() async {
    return true;
  }

  static Future<bool> checkAndRequestPermission() async {
    if (await checkPermission()) {
      return true;
    }
    return await requestPermission();
  }

  static Future<bool> isLocationServiceEnabled() async {
    return true;
  }

  static Future<void> openSettings() async {
    // Desktop: no-op
  }
}

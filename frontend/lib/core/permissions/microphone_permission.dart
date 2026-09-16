import 'dart:io' show Platform;

class MicrophonePermission {
  static Future<bool> checkPermission() async {
    if (Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
      return true;
    }
    return true;
  }

  static Future<bool> requestPermission() async {
    if (Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
      return true;
    }
    return true;
  }

  static Future<bool> checkAndRequestPermission() async {
    if (await checkPermission()) {
      return true;
    }
    return await requestPermission();
  }

  static Future<void> openSettings() async {
    // Desktop: no-op
  }
}

import 'package:permission_handler/permission_handler.dart';

class MicrophonePermission {
  static Future<bool> checkPermission() async {
    final status = await Permission.microphone.status;
    return status.isGranted;
  }

  static Future<bool> requestPermission() async {
    final status = await Permission.microphone.request();
    return status.isGranted;
  }

  static Future<bool> checkAndRequestPermission() async {
    if (await checkPermission()) {
      return true;
    }
    return await requestPermission();
  }

  static Future<void> openSettings() async {
    await openAppSettings();
  }
}

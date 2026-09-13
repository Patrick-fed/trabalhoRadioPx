import 'package:permission_handler/permission_handler.dart';

class LocationPermissionHelper {
  static Future<bool> checkPermission() async {
    final status = await Permission.locationWhenInUse.status;
    return status.isGranted;
  }

  static Future<bool> requestPermission() async {
    final status = await Permission.locationWhenInUse.request();
    return status.isGranted;
  }

  static Future<bool> checkAndRequestPermission() async {
    if (await checkPermission()) {
      return true;
    }
    return await requestPermission();
  }

  static Future<bool> isLocationServiceEnabled() async {
    return await Permission.locationWhenInUse.serviceStatus.isEnabled;
  }

  static Future<void> openSettings() async {
    await openAppSettings();
  }
}

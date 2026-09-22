import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:record/record.dart';

class MicrophonePermission {
  static bool _hasRequested = false;

  static Future<bool> checkPermission() async {
    if (kIsWeb) return true;

    final recorder = AudioRecorder();
    try {
      return await recorder.hasPermission();
    } catch (_) {
      return false;
    } finally {
      recorder.dispose();
    }
  }

  static Future<bool> requestPermission() async {
    if (kIsWeb) return true;

    final recorder = AudioRecorder();
    try {
      return await recorder.hasPermission();
    } catch (_) {
      return false;
    } finally {
      recorder.dispose();
    }
  }

  static Future<bool> checkAndRequestPermission() async {
    final granted = await checkPermission();
    if (granted) return true;
    _hasRequested = true;
    return await requestPermission();
  }

  static bool get hasRequested => _hasRequested;
}
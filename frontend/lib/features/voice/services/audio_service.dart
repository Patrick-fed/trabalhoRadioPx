import 'dart:async';
import 'dart:io' show Platform;
import 'dart:typed_data';

class AudioService {
  bool _isRecording = false;
  StreamController<Uint8List>? _audioStreamController;

  bool get isRecording => _isRecording;
  bool get _isDesktop => Platform.isWindows || Platform.isLinux || Platform.isMacOS;

  Future<bool> checkPermission() async {
    return true;
  }

  Future<void> startRecording({
    required void Function(Uint8List audio) onData,
  }) async {
    if (_isRecording) return;

    _audioStreamController = StreamController<Uint8List>.broadcast();
    _isRecording = true;

    if (!_isDesktop) {
      // TODO: Implement real recording for mobile using platform channels
    }
  }

  Future<void> stopRecording() async {
    if (!_isRecording) return;

    _isRecording = false;
    _audioStreamController?.close();
    _audioStreamController = null;
  }

  Stream<Uint8List>? get audioStream => _audioStreamController?.stream;

  void dispose() {
    stopRecording();
    _audioStreamController?.close();
  }
}

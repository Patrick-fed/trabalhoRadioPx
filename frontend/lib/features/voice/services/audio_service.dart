import 'dart:async';
import 'dart:typed_data';

import 'package:record/record.dart';

import '../../../core/permissions/microphone_permission.dart';

class AudioService {
  final AudioRecorder _recorder = AudioRecorder();
  StreamSubscription<Uint8List>? _streamSubscription;
  bool _isRecording = false;

  bool get isRecording => _isRecording;

  Future<bool> checkPermission() async {
    return MicrophonePermission.checkAndRequestPermission();
  }

  Future<void> startRecording({
    required void Function(Uint8List audio) onData,
    void Function()? onError,
  }) async {
    if (_isRecording) return;

    if (!await checkPermission()) {
      onError?.call();
      return;
    }

    const config = RecordConfig(
      encoder: AudioEncoder.pcm16bits,
      sampleRate: 16000,
      numChannels: 1,
      autoGain: true,
      echoCancel: false,
    );

    try {
      final stream = await _recorder.startStream(config);
      _streamSubscription = stream.listen(
        onData,
        onError: (Object e) {
          onError?.call();
        },
      );
      _isRecording = true;
    } catch (_) {
      onError?.call();
    }
  }

  Future<void> stopRecording() async {
    if (!_isRecording) return;

    _isRecording = false;
    await _streamSubscription?.cancel();
    _streamSubscription = null;

    try {
      await _recorder.stop();
    } catch (_) {
      // Stream already stopped.
    }
  }

  void dispose() {
    if (_isRecording) {
      stopRecording();
    }
    _recorder.dispose();
  }
}
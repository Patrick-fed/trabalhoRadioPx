import 'dart:async';
import 'dart:typed_data';

import 'package:record/record.dart';

class AudioService {
  final AudioRecorder _recorder = AudioRecorder();
  bool _isRecording = false;
  StreamController<Uint8List>? _audioStreamController;

  bool get isRecording => _isRecording;

  Future<bool> checkPermission() async {
    return await _recorder.hasPermission();
  }

  Future<void> startRecording({
    required void Function(Uint8List audio) onData,
  }) async {
    if (_isRecording) return;

    final hasPermission = await checkPermission();
    if (!hasPermission) {
      throw Exception('Microphone permission not granted');
    }

    _audioStreamController = StreamController<Uint8List>.broadcast();

    await _recorder.start(
      RecordConfig(
        encoder: AudioEncoder.opus,
        numChannels: 1,
        sampleRate: 48000,
      ),
      onFreshData: (data) {
        onData(data);
      },
    );

    _isRecording = true;
  }

  Future<void> stopRecording() async {
    if (!_isRecording) return;

    await _recorder.stop();
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

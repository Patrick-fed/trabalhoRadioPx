import 'dart:async';
import 'dart:typed_data';

import 'audio_service.dart';

class PttService {
  final AudioService _audioService;
  bool _isTransmitting = false;
  String? _currentChannelId;
  String? _userId;

  PttService(this._audioService);

  bool get isTransmitting => _isTransmitting;
  String? get currentChannelId => _currentChannelId;

  Future<void> startTransmitting({
    required String channelId,
    required String userId,
    required void Function(Uint8List audio) onAudioData,
  }) async {
    if (_isTransmitting) return;

    _currentChannelId = channelId;
    _userId = userId;
    _isTransmitting = true;

    await _audioService.startRecording(
      onData: (data) {
        onAudioData(data);
      },
    );
  }

  Future<void> stopTransmitting() async {
    if (!_isTransmitting) return;

    await _audioService.stopRecording();
    _isTransmitting = false;
    _currentChannelId = null;
    _userId = null;
  }

  void dispose() {
    stopTransmitting();
  }
}

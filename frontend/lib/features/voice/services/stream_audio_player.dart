import 'dart:typed_data';

import 'package:flutter_sound/flutter_sound.dart';

class StreamAudioPlayer {
  static const int _sampleRate = 16000;
  static const int _numChannels = 1;

  final FlutterSoundPlayer _player = FlutterSoundPlayer();
  bool _inSegment = false;
  bool _sessionReady = false;

  bool get inSegment => _inSegment;

  Future<void> startSession() async {
    if (_sessionReady) return;

    await _player.openPlayer();
    await _player.startPlayerFromStream(
      codec: Codec.pcm16,
      interleaved: true,
      numChannels: _numChannels,
      sampleRate: _sampleRate,
      bufferSize: 2048,
    );

    _sessionReady = true;
  }

  void startSegment() {
    if (_inSegment) {
      return;
    }
    _inSegment = true;
  }

  void feed(Uint8List pcmData) {
    if (!_sessionReady || !_inSegment) {
      return;
    }

    final sink = _player.uint8ListSink;
    if (sink == null) return;

    try {
      sink.add(pcmData);
    } catch (_) {
      // Player underflow ou sink fechado: ignora e aguarda o próximo pacote.
    }
  }

  void endOfSegment() {
    _inSegment = false;
  }

  Future<void> dispose() async {
    endOfSegment();
    _sessionReady = false;
    await _player.stopPlayer();
    await _player.closePlayer();
  }
}
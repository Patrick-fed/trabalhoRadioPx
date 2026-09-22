import 'dart:async';
import 'dart:typed_data';

import 'package:just_audio/just_audio.dart';

class StreamAudioPlayer {
  static const int _sampleRate = 16000;
  static const int _numChannels = 1;
  static const int _bitsPerSample = 16;

  final AudioPlayer _player = AudioPlayer();
  StreamController<Uint8List>? _streamController;
  StreamAudioSource? _source;
  bool _inSegment = false;
  bool _headerSent = false;

  bool get inSegment => _inSegment;

  void startSegment() {
    if (_inSegment) {
      return;
    }

    _streamController = StreamController<Uint8List>();
    _source = _PcmStreamAudioSource(_streamController!.stream);

    _inSegment = true;
    _headerSent = false;

    _player
        .setAudioSource(_source!, initialPosition: Duration.zero)
        .then((_) => _player.play())
        .catchError((_) {});
  }

  void feed(Uint8List pcmData) {
    final controller = _streamController;
    if (!_inSegment || controller == null || controller.isClosed) {
      return;
    }

    if (!_headerSent) {
      controller.add(buildWavHeader(
        sampleRate: _sampleRate,
        numChannels: _numChannels,
        bitsPerSample: _bitsPerSample,
      ));
      _headerSent = true;
    }

    controller.add(pcmData);
  }

  void endOfSegment() {
    if (!_inSegment) {
      return;
    }

    _inSegment = false;
    _headerSent = false;

    final controller = _streamController;
    _streamController = null;
    controller?.close();

    _player.stop().catchError((_) {});
  }

  Future<void> dispose() async {
    endOfSegment();
    await _player.dispose();
  }
}

class _PcmStreamAudioSource extends StreamAudioSource {
  final Stream<Uint8List> _stream;

  _PcmStreamAudioSource(this._stream) : super(tag: 'radiopx-live');

  @override
  Future<StreamAudioResponse> request([int? start, int? end]) async {
    return StreamAudioResponse(
      sourceLength: null,
      contentLength: null,
      offset: start,
      contentType: 'audio/wav',
      stream: _stream,
    );
  }
}

Uint8List buildWavHeader({
  required int sampleRate,
  required int numChannels,
  required int bitsPerSample,
}) {
  final blockAlign = numChannels * bitsPerSample ~/ 8;
  final byteRate = sampleRate * blockAlign;
  final dataSize = 0xFFFFFFFF;

  final bytes = ByteData(44);
  void writeAscii(int offset, String value) {
    for (var i = 0; i < value.length; i++) {
      bytes.setUint8(offset + i, value.codeUnitAt(i));
    }
  }

  writeAscii(0, 'RIFF');
  bytes.setUint32(4, 36 + dataSize, Endian.little);
  writeAscii(8, 'WAVE');
  writeAscii(12, 'fmt ');
  bytes.setUint32(16, 16, Endian.little);
  bytes.setUint16(20, 1, Endian.little);
  bytes.setUint16(22, numChannels, Endian.little);
  bytes.setUint32(24, sampleRate, Endian.little);
  bytes.setUint32(28, byteRate, Endian.little);
  bytes.setUint16(32, blockAlign, Endian.little);
  bytes.setUint16(34, bitsPerSample, Endian.little);
  writeAscii(36, 'data');
  bytes.setUint32(40, dataSize, Endian.little);

  return bytes.buffer.asUint8List();
}
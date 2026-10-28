import 'dart:async';
import 'dart:collection';
import 'dart:typed_data';

/// FIFO de chunks de áudio com teto de memória de 30s (RF08) e sentinel de
/// fim de segmento (RF06).
class AudioQueue {
  /// 30s de PCM16 mono 16 kHz ≈ 960.000 bytes.
  static const int defaultMaxBytes = 960000;

  final int _maxBytes;
  final Queue<_AudioChunk> _queue = Queue<_AudioChunk>();
  int _totalBytes = 0;
  bool _isPlaying = false;
  Timer? _playTimer;
  final Duration _playbackInterval = const Duration(milliseconds: 50);

  final StreamController<bool> _playingController =
      StreamController<bool>.broadcast();
  final StreamController<Uint8List> _audioController =
      StreamController<Uint8List>.broadcast();
  final StreamController<void> _endOfSegmentController =
      StreamController<void>.broadcast();

  AudioQueue({int maxBytes = defaultMaxBytes}) : _maxBytes = maxBytes;

  Stream<bool> get isPlayingStream => _playingController.stream;
  Stream<Uint8List> get audioStream => _audioController.stream;
  Stream<void> get endOfSegmentStream => _endOfSegmentController.stream;
  bool get isPlaying => _isPlaying;
  int get queueLength => _queue.length;
  int get queuedBytes => _totalBytes;

  void enqueue(Uint8List audioData, {String? senderId}) {
    if (audioData.isEmpty) return;

    _queue.add(_AudioChunk(
      data: audioData,
      senderId: senderId,
      timestamp: DateTime.now(),
    ));
    _totalBytes += audioData.length;

    // Teto de 30s: descarta os chunks mais antigos (nunca o último).
    while (_totalBytes > _maxBytes && _queue.length > 1) {
      final dropped = _queue.removeFirst();
      if (!dropped.isEndOfSegment) {
        _totalBytes -= dropped.data.length;
      }
    }
  }

  void enqueueEndOfSegment({String? senderId}) {
    _queue.add(_AudioChunk(
      data: Uint8List(0),
      senderId: senderId,
      timestamp: DateTime.now(),
      isEndOfSegment: true,
    ));
  }

  Future<void> startPlayback() async {
    if (_isPlaying) return;

    _isPlaying = true;
    _playingController.add(true);

    _playTimer = Timer.periodic(_playbackInterval, (_) {
      _processNextChunk();
    });
  }

  Future<void> stopPlayback() async {
    _isPlaying = false;
    _playingController.add(false);
    _playTimer?.cancel();
    _playTimer = null;
  }

  void clear() {
    _queue.clear();
    _totalBytes = 0;
    stopPlayback();
  }

  void _processNextChunk() {
    if (_queue.isEmpty) {
      stopPlayback();
      return;
    }

    final chunk = _queue.removeFirst();
    if (chunk.isEndOfSegment) {
      _endOfSegmentController.add(null);
      return;
    }

    _totalBytes -= chunk.data.length;
    _audioController.add(chunk.data);
  }

  void dispose() {
    clear();
    _playingController.close();
    _audioController.close();
    _endOfSegmentController.close();
  }
}

class _AudioChunk {
  final Uint8List data;
  final String? senderId;
  final DateTime timestamp;
  final bool isEndOfSegment;

  _AudioChunk({
    required this.data,
    this.senderId,
    required this.timestamp,
    this.isEndOfSegment = false,
  });
}

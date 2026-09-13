import 'dart:async';
import 'dart:collection';
import 'dart:typed_data';

class AudioQueue {
  final Queue<_AudioChunk> _queue = Queue<_AudioChunk>();
  bool _isPlaying = false;
  Timer? _playTimer;
  Duration _playbackInterval = const Duration(milliseconds: 50);

  final StreamController<bool> _playingController =
      StreamController<bool>.broadcast();
  final StreamController<Uint8List> _audioController =
      StreamController<Uint8List>.broadcast();

  Stream<bool> get isPlayingStream => _playingController.stream;
  Stream<Uint8List> get audioStream => _audioController.stream;
  bool get isPlaying => _isPlaying;
  int get queueLength => _queue.length;

  void enqueue(Uint8List audioData, {String? senderId}) {
    _queue.add(_AudioChunk(
      data: audioData,
      senderId: senderId,
      timestamp: DateTime.now(),
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
    stopPlayback();
  }

  void _processNextChunk() {
    if (_queue.isEmpty) {
      stopPlayback();
      return;
    }

    final chunk = _queue.removeFirst();
    _audioController.add(chunk.data);
  }

  void dispose() {
    clear();
    _playingController.close();
    _audioController.close();
  }
}

class _AudioChunk {
  final Uint8List data;
  final String? senderId;
  final DateTime timestamp;

  _AudioChunk({
    required this.data,
    this.senderId,
    required this.timestamp,
  });
}

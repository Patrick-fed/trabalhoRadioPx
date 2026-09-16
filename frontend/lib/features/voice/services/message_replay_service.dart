import 'dart:async';
import 'dart:collection';
import 'dart:typed_data';

import '../models/message_buffer_model.dart';
import '../../../core/websocket/websocket_manager.dart';

class MessageReplayService {
  final WebSocketManager _webSocketManager;
  final AudioQueue _audioQueue;
  
  final Map<String, List<MessageBufferModel>> _localBuffer = {};
  final Map<String, int> _lastSequenceMap = {};
  
  bool _isReplaying = false;
  final StreamController<bool> _replayController =
      StreamController<bool>.broadcast();

  Stream<bool> get isReplayingStream => _replayController.stream;
  bool get isReplaying => _isReplaying;

  MessageReplayService(this._webSocketManager, this._audioQueue) {
    _webSocketManager.connectionStream.listen(_onConnectionChange);
  }

  void _onConnectionChange(bool isConnected) {
    if (isConnected) {
      _handleReconnection();
    } else {
      _handleDisconnection();
    }
  }

  void _handleDisconnection() {
    // Mark current state for replay
  }

  void _handleReconnection() async {
    if (_isReplaying) return;

    _isReplaying = true;
    _replayController.add(true);

    try {
      await _requestPendingMessages();
    } catch (e) {
      print('Error during replay: $e');
    } finally {
      _isReplaying = false;
      _replayController.add(false);
    }
  }

  Future<void> _requestPendingMessages() async {
    final messages = _getUnreplayedMessages();
    
    if (messages.isEmpty) return;

    for (final message in messages) {
      await _replayMessage(message);
      _markAsReplayed(message.id);
    }

    await _confirmReplay(messages.map((m) => m.id).toList());
  }

  List<MessageBufferModel> _getUnreplayedMessages() {
    final allMessages = <MessageBufferModel>[];
    
    for (final messages in _localBuffer.values) {
      for (final message in messages) {
        if (!message.isReplayed) {
          allMessages.add(message);
        }
      }
    }

    allMessages.sort((a, b) => a.sequence.compareTo(b.sequence));
    
    return allMessages;
  }

  Future<void> _replayMessage(MessageBufferModel message) async {
    await _audioQueue.enqueue(message.audioData);
  }

  void _markAsReplayed(String messageId) {
    for (final channelId in _localBuffer.keys) {
      final messages = _localBuffer[channelId];
      if (messages == null) continue;
      for (int i = 0; i < messages.length; i++) {
        if (messages[i].id == messageId) {
          messages[i] = messages[i].copyWith(isReplayed: true);
          break;
        }
      }
    }
  }

  Future<void> _confirmReplay(List<String> messageIds) async {
    // In a real implementation, this would call the backend API
    // to confirm that messages have been replayed
  }

  void addToBuffer(String channelId, MessageBufferModel message) {
    if (!_localBuffer.containsKey(channelId)) {
      _localBuffer[channelId] = [];
    }

    _localBuffer[channelId]!.add(message);
    _cleanupOldMessages(channelId);
  }

  void _cleanupOldMessages(String channelId) {
    final messages = _localBuffer[channelId];
    if (messages == null) return;

    final now = DateTime.now();
    final cutoff = now.subtract(const Duration(seconds: 30));

    messages.removeWhere((msg) => msg.timestamp.isBefore(cutoff));
  }

  void clearBuffer(String channelId) {
    _localBuffer.remove(channelId);
    _lastSequenceMap.remove(channelId);
  }

  void clearAllBuffers() {
    _localBuffer.clear();
    _lastSequenceMap.clear();
  }

  void dispose() {
    _replayController.close();
    _localBuffer.clear();
    _lastSequenceMap.clear();
  }
}

class AudioQueue {
  final Queue<Uint8List> _queue = Queue<Uint8List>();
  bool _isPlaying = false;
  Timer? _playTimer;
  
  final StreamController<Uint8List> _audioController =
      StreamController<Uint8List>.broadcast();

  Stream<Uint8List> get audioStream => _audioController.stream;
  bool get isPlaying => _isPlaying;
  int get queueLength => _queue.length;

  Future<void> enqueue(Uint8List audioData) async {
    _queue.add(audioData);
    
    if (!_isPlaying) {
      await startPlayback();
    }
  }

  Future<void> startPlayback() async {
    if (_isPlaying) return;

    _isPlaying = true;
    _playTimer = Timer.periodic(
      const Duration(milliseconds: 50),
      (_) => _processNextChunk(),
    );
  }

  Future<void> stopPlayback() async {
    _isPlaying = false;
    _playTimer?.cancel();
    _playTimer = null;
  }

  void _processNextChunk() {
    if (_queue.isEmpty) {
      stopPlayback();
      return;
    }

    final chunk = _queue.removeFirst();
    _audioController.add(chunk);
  }

  void clear() {
    _queue.clear();
    stopPlayback();
  }

  void dispose() {
    clear();
    _audioController.close();
  }
}

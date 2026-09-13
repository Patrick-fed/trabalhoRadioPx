import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class WebSocketService {
  WebSocketChannel? _channel;
  final StreamController<Map<String, dynamic>> _messageController =
      StreamController<Map<String, dynamic>>.broadcast();
  bool _isConnected = false;
  String? _currentChannelId;

  Stream<Map<String, dynamic>> get messageStream => _messageController.stream;
  bool get isConnected => _isConnected;
  String? get currentChannelId => _currentChannelId;

  Future<void> connect({
    required String channelId,
    required String userId,
    String baseUrl = 'ws://localhost:8080',
  }) async {
    if (_isConnected) {
      await disconnect();
    }

    _currentChannelId = channelId;
    final url = '$baseUrl/ws/audio?channel=$channelId&user_id=$userId';

    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));

      _channel!.stream.listen(
        (data) {
          final message = jsonDecode(data as String) as Map<String, dynamic>;
          _messageController.add(message);
        },
        onDone: () {
          _isConnected = false;
          _messageController.add({'type': 'disconnected'});
        },
        onError: (error) {
          _isConnected = false;
          _messageController.add({'type': 'error', 'error': error.toString()});
        },
      );

      _isConnected = true;
      _messageController.add({'type': 'connected'});
    } catch (e) {
      _isConnected = false;
      _messageController.add({'type': 'error', 'error': e.toString()});
    }
  }

  Future<void> disconnect() async {
    await _channel?.sink.close();
    _channel = null;
    _isConnected = false;
    _currentChannelId = null;
  }

  void sendAudio(List<int> audioData) {
    if (!_isConnected || _channel == null) return;

    final message = jsonEncode({
      'type': 'audio',
      'payload': {
        'audio': audioData,
        'is_ptt': true,
        'channel_id': _currentChannelId,
      },
    });

    _channel!.sink.add(message);
  }

  void sendJoinRoom(String roomId) {
    if (!_isConnected || _channel == null) return;

    final message = jsonEncode({
      'type': 'join_room',
      'payload': {'room': roomId},
    });

    _channel!.sink.add(message);
  }

  void sendLeaveRoom(String roomId) {
    if (!_isConnected || _channel == null) return;

    final message = jsonEncode({
      'type': 'leave_room',
      'payload': {'room': roomId},
    });

    _channel!.sink.add(message);
  }

  void dispose() {
    disconnect();
    _messageController.close();
  }
}

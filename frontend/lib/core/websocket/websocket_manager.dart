import 'dart:async';

import 'package:web_socket_channel/web_socket_channel.dart';

class WebSocketManager {
  WebSocketChannel? _channel;
  final StreamController<Map<String, dynamic>> _messageController =
      StreamController<Map<String, dynamic>>.broadcast();
  final StreamController<bool> _connectionController =
      StreamController<bool>.broadcast();
  
  bool _isConnected = false;
  String? _currentChannelId;
  String? _userId;
  Timer? _reconnectTimer;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 5;
  static const Duration _reconnectDelay = Duration(seconds: 2);

  Stream<Map<String, dynamic>> get messageStream => _messageController.stream;
  Stream<bool> get connectionStream => _connectionController.stream;
  bool get isConnected => _isConnected;
  String? get currentChannelId => _currentChannelId;

  Future<void> connect({
    required String channelId,
    required String userId,
    String baseUrl = 'ws://localhost:8080',
  }) async {
    _currentChannelId = channelId;
    _userId = userId;
    
    await _establishConnection(baseUrl);
  }

  Future<void> _establishConnection(String baseUrl) async {
    if (_isConnected) {
      await disconnect();
    }

    final url = '$baseUrl/ws/audio?channel=$_currentChannelId&user_id=$_userId';

    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));

      _channel!.stream.listen(
        (data) {
          final message = _parseMessage(data);
          if (message != null) {
            _messageController.add(message);
          }
        },
        onDone: () {
          _handleDisconnection();
        },
        onError: (error) {
          _handleDisconnection();
        },
      );

      _isConnected = true;
      _reconnectAttempts = 0;
      _connectionController.add(true);
      _messageController.add({'type': 'connected'});
    } catch (e) {
      _isConnected = false;
      _connectionController.add(false);
      _messageController.add({'type': 'error', 'error': e.toString()});
    }
  }

  void _handleDisconnection() {
    _isConnected = false;
    _connectionController.add(false);
    _messageController.add({'type': 'disconnected'});
    
    _attemptReconnect();
  }

  void _attemptReconnect() {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      _messageController.add({
        'type': 'reconnect_failed',
        'message': 'Max reconnection attempts reached',
      });
      return;
    }

    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(_reconnectDelay, () async {
      _reconnectAttempts++;
      _messageController.add({
        'type': 'reconnecting',
        'attempt': _reconnectAttempts,
      });
      
      await _establishConnection('ws://localhost:8080');
    });
  }

  Map<String, dynamic>? _parseMessage(dynamic data) {
    try {
      if (data is String) {
        return Map<String, dynamic>.from(
          Map.fromEntries(
            (data as Map).entries.map(
              (e) => MapEntry(e.key.toString(), e.value),
            ),
          ),
        );
      }
      return null;
    } catch (e) {
      return null;
    }
  }

  Future<void> disconnect() async {
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    
    await _channel?.sink.close();
    _channel = null;
    _isConnected = false;
    _currentChannelId = null;
    _userId = null;
    _reconnectAttempts = 0;
  }

  void sendAudio(List<int> audioData) {
    if (!_isConnected || _channel == null) return;

    final message = {
      'type': 'audio',
      'payload': {
        'audio': audioData,
        'is_ptt': true,
        'channel_id': _currentChannelId,
      },
    };

    _channel!.sink.add(message);
  }

  void sendJoinRoom(String roomId) {
    if (!_isConnected || _channel == null) return;

    final message = {
      'type': 'join_room',
      'payload': {'room': roomId},
    };

    _channel!.sink.add(message);
  }

  void sendLeaveRoom(String roomId) {
    if (!_isConnected || _channel == null) return;

    final message = {
      'type': 'leave_room',
      'payload': {'room': roomId},
    };

    _channel!.sink.add(message);
  }

  void dispose() {
    disconnect();
    _messageController.close();
    _connectionController.close();
  }
}

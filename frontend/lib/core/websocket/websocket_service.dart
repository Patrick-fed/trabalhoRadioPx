import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../api_client.dart';
import '../config.dart';

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
    required String token,
  }) async {
    if (_isConnected) {
      await disconnect();
    }

    _currentChannelId = channelId;
    final uri = Uri.parse('${AppConfig.wsBaseUrl}/ws/audio').replace(
      queryParameters: {
        'channel': channelId,
        'user_id': userId,
        'token': token,
      },
    );

    final channel = WebSocketChannel.connect(uri);

    try {
      await channel.ready.timeout(apiTimeout);
    } catch (_) {
      // O sink só recebe destino quando o handshake conclui; fechar sem
      // aguardar evita pendurar quando a conexão nunca se estabelece.
      unawaited(channel.sink.close().timeout(
        const Duration(seconds: 2),
        onTimeout: () {},
      ));
      _currentChannelId = null;
      throw Exception('Não foi possível conectar ao servidor de áudio.');
    }

    _channel = channel;

    channel.stream.listen(
      (data) {
        try {
          final decoded = jsonDecode(data as String);
          if (decoded is Map<String, dynamic>) {
            _messageController.add(decoded);
          }
        } catch (_) {
          _messageController
              .add({'type': 'error', 'error': 'Invalid message received'});
        }
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
  }

  Future<void> disconnect() async {
    await _channel?.sink.close();
    _channel = null;
    _isConnected = false;
    _currentChannelId = null;
  }

  void sendAudio(List<int> audioData, {bool isPtt = true}) {
    if (!_isConnected || _channel == null) return;

    final message = jsonEncode({
      'type': 'audio',
      'payload': {
        'room': _currentChannelId,
        'audio': base64Encode(audioData),
        'is_ptt': isPtt,
      },
    });

    _channel!.sink.add(message);
  }

  /// RF06: pede ao servidor o áudio bufferizado desde a última sequence
  /// conhecida (replay por usuário após reconexão).
  void requestReplay(String roomId, int sinceSequence) {
    if (!_isConnected || _channel == null) return;

    final message = jsonEncode({
      'type': 'request_replay',
      'payload': {
        'room': roomId,
        'since_sequence': sinceSequence,
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
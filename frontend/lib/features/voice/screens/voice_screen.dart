import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';

import '../../../core/websocket/websocket_service.dart';
import '../../auth/services/auth_service.dart';
import '../../channels/services/channel_service.dart';
import '../services/audio_queue.dart';
import '../services/audio_service.dart';
import '../services/stream_audio_player.dart';
import '../widgets/audio_player.dart';
import '../widgets/audio_visualizer.dart';
import '../widgets/ptt_button.dart';

class VoiceScreen extends StatefulWidget {
  final String channelId;
  final String channelName;
  final VoidCallback? onLeaveChannel;

  const VoiceScreen({
    super.key,
    required this.channelId,
    required this.channelName,
    this.onLeaveChannel,
  });

  @override
  State<VoiceScreen> createState() => _VoiceScreenState();
}

class _VoiceScreenState extends State<VoiceScreen> {
  final WebSocketService _webSocketService = WebSocketService();
  final AudioService _audioService = AudioService();
  final StreamAudioPlayer _streamPlayer = StreamAudioPlayer();
  final AuthService _authService = AuthService();
  final ChannelService _channelService = ChannelService();
  final AudioQueue _audioQueue = AudioQueue();

  bool _isTransmitting = false;
  bool _isReceiving = false;
  bool _isChannelBusy = false;
  bool _connectionError = false;
  bool _connectionFailed = false;
  String? _currentSpeaker;
  String? _userId;
  StreamSubscription<Map<String, dynamic>>? _wsSubscription;
  StreamSubscription<Uint8List>? _audioSub;
  StreamSubscription<void>? _endOfSegmentSub;
  Timer? _reconnectTimer;
  Timer? _receiveIdleTimer;
  int _reconnectAttempts = 0;

  /// Última sequence de áudio recebida (RF06: base do replay pós-reconexão).
  int _lastSequence = 0;

  static const Duration _receiveIdleTimeout = Duration(milliseconds: 2500);
  static const int _maxReconnectAttempts = 5;
  static const List<Duration> _reconnectBackoff = [
    Duration(seconds: 2),
    Duration(seconds: 4),
    Duration(seconds: 8),
    Duration(seconds: 16),
    Duration(seconds: 30),
  ];

  @override
  void initState() {
    super.initState();
    _init();
  }

  Future<void> _init() async {
    final user = await _authService.getCurrentUser();
    final uid = user?.id ?? 'anonymous';
    _userId = uid;

    try {
      await _streamPlayer.startSession();
    } catch (_) {
      // Sessão de áudio indisponível: segue sem reprodução de som.
    }

    _audioSub = _audioQueue.audioStream.listen(_onQueueAudio);
    _endOfSegmentSub =
        _audioQueue.endOfSegmentStream.listen((_) => _endReceiveSegment());

    await _connectWebSocket();
  }

  Future<void> _connectWebSocket() async {
    if (_webSocketService.isConnected) return;

    _wsSubscription?.cancel();
    await _webSocketService.disconnect();

    final token = await _authService.getValidToken();
    if (token == null) {
      _scheduleReconnect();
      return;
    }

    try {
      await _webSocketService.connect(
        channelId: widget.channelId,
        userId: _userId ?? 'anonymous',
        token: token,
      );
    } catch (_) {
      if (mounted) _scheduleReconnect();
      return;
    }

    _wsSubscription =
        _webSocketService.messageStream.listen(_onWebSocketMessage);

    // RF06: reconexão com histórico — pede o áudio perdido desde a última
    // sequence (na primeira conexão _lastSequence é 0 e nada é solicitado).
    if (_lastSequence > 0) {
      _webSocketService.requestReplay(widget.channelId, _lastSequence);
    }

    if (mounted) {
      setState(() {
        _connectionError = false;
        _connectionFailed = false;
        _reconnectAttempts = 0;
      });
    }
  }

  void _onWebSocketMessage(Map<String, dynamic> message) {
    switch (message['type']) {
      case 'audio':
        _handleAudioMessage(message);
        break;
      case 'disconnected':
        // Zera o estado de recepção para o canal não ficar "ocupado" para
        // sempre; a fila e a sequence são preservadas para o replay.
        _endReceiveSegment();
        _scheduleReconnect();
        break;
      case 'error':
        if (mounted) {
          setState(() {
            _connectionError = true;
          });
        }
        break;
    }
  }

  void _handleAudioMessage(Map<String, dynamic> message) {
    final payload = message['payload'];
    if (payload is! Map) return;

    final isPtt = payload['is_ptt'] == true;
    final audioRaw = payload['audio'];
    final sequence = payload['sequence'];
    if (sequence is int && sequence > _lastSequence) {
      _lastSequence = sequence;
    }

    if (audioRaw is String && audioRaw.isNotEmpty) {
      _audioQueue.enqueue(
        Uint8List.fromList(base64Decode(audioRaw)),
        senderId: payload['user_id'] as String?,
      );
      _audioQueue.startPlayback();
    }

    if (!isPtt) {
      _audioQueue.enqueueEndOfSegment();
    }
  }

  void _onQueueAudio(Uint8List bytes) {
    if (!_streamPlayer.inSegment) {
      if (mounted) {
        setState(() {
          _isReceiving = true;
          _isChannelBusy = true;
        });
      }
      _streamPlayer.startSegment();
    }

    _streamPlayer.feed(bytes);
    _resetReceiveIdleTimer();
  }

  void _endReceiveSegment() {
    _receiveIdleTimer?.cancel();
    _receiveIdleTimer = null;
    _streamPlayer.endOfSegment();
    if (mounted) {
      setState(() {
        _isReceiving = false;
        if (!_isTransmitting) {
          _isChannelBusy = false;
        }
      });
    }
  }

  /// Evita travar "recebendo" se o falante cair sem enviar is_ptt=false.
  void _resetReceiveIdleTimer() {
    _receiveIdleTimer?.cancel();
    _receiveIdleTimer = Timer(_receiveIdleTimeout, _endReceiveSegment);
  }

  void _scheduleReconnect() {
    if (!mounted || _reconnectTimer != null) return;

    if (_reconnectAttempts >= _maxReconnectAttempts) {
      setState(() {
        _connectionFailed = true;
        _connectionError = true;
      });
      return;
    }

    final delay = _reconnectBackoff[_reconnectAttempts];
    _reconnectAttempts++;

    setState(() {
      _connectionError = true;
    });

    _reconnectTimer = Timer(delay, () {
      _reconnectTimer = null;
      if (!mounted) return;
      _connectWebSocket();
    });
  }

  void _retryConnection() {
    if (!mounted || _reconnectTimer != null) return;

    setState(() {
      _connectionFailed = false;
      _connectionError = false;
      _reconnectAttempts = 0;
    });

    _connectWebSocket();
  }

  Future<void> _startTalking() async {
    if (_isTransmitting) return;

    final hasPermission = await _audioService.checkPermission();
    if (!hasPermission) {
      _showSnack('Permissão de microfone negada');
      return;
    }

    await _audioService.startRecording(
      onData: (Uint8List chunk) {
        if (_webSocketService.isConnected) {
          _webSocketService.sendAudio(chunk);
        }
      },
      onError: () => _showSnack('Erro ao capturar áudio'),
    );

    if (!_audioService.isRecording) {
      _showSnack('Não foi possível iniciar a gravação');
      return;
    }

    if (_webSocketService.isConnected) {
      _webSocketService.sendJoinRoom(widget.channelId);
    }

    if (!mounted) return;
    setState(() {
      _isTransmitting = true;
      _isChannelBusy = true;
      _currentSpeaker = 'Você';
    });
  }

  Future<void> _stopTalking() async {
    await _audioService.stopRecording();
    _webSocketService.sendAudio(const [], isPtt: false);

    if (!mounted) return;
    setState(() {
      _isTransmitting = false;
      _currentSpeaker = null;
      if (!_isReceiving) {
        _isChannelBusy = false;
      }
    });
  }

  Future<void> _leaveChannel() async {
    try {
      await _channelService.leaveChannel(widget.channelId);
    } catch (_) {
      // Ignore errors on leave.
    }
    if (mounted) {
      widget.onLeaveChannel?.call();
    }
  }

  Future<void> _showChannelUsers() async {
    try {
      final users = await _channelService.getChannelUsers(widget.channelId);
      if (!mounted) return;
      showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Membros do Canal'),
          content: SizedBox(
            width: double.maxFinite,
            child: users.isEmpty
                ? const Text('Nenhum membro')
                : ListView.builder(
                    shrinkWrap: true,
                    itemCount: users.length,
                    itemBuilder: (context, index) => ListTile(
                      dense: true,
                      leading: const Icon(Icons.person),
                      title: Text(users[index]),
                    ),
                  ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Fechar'),
            ),
          ],
        ),
      );
    } catch (_) {
      _showSnack('Não foi possível carregar os membros');
    }
  }

  void _showSnack(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  void dispose() {
    _reconnectTimer?.cancel();
    _receiveIdleTimer?.cancel();
    _wsSubscription?.cancel();
    _audioSub?.cancel();
    _endOfSegmentSub?.cancel();
    _audioQueue.dispose();
    _audioService.dispose();
    _streamPlayer.dispose();
    _webSocketService.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_currentChannelTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: _showChannelSettings,
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (_connectionError)
              Container(
                margin: const EdgeInsets.only(bottom: 16),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.wifi_off, color: Colors.red, size: 18),
                    const SizedBox(width: 8),
                    Text(
                      _connectionFailed
                          ? 'Sem conexão com o servidor'
                          : 'Conexão perdida. Reconectando...',
                      style: const TextStyle(
                        color: Colors.red,
                        fontSize: 13,
                      ),
                    ),
                  ],
                ),
              ),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.blue.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.radio, color: Colors.blue),
                  const SizedBox(width: 8),
                  Text(
                    widget.channelName,
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 32),

            if (_isTransmitting) ...[
              const AudioVisualizer(
                isActive: true,
                color: Colors.red,
                height: 80,
                barCount: 7,
              ),
              const SizedBox(height: 8),
              const Text(
                'TRANSMITINDO',
                style: TextStyle(
                  color: Colors.red,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
            ] else if (_isReceiving) ...[
              AudioPlayerWidget(
                isPlaying: true,
                currentSpeaker: _currentSpeaker,
                onStop: () {
                  _audioQueue.clear();
                  _endReceiveSegment();
                },
              ),
            ] else ...[
              const AudioVisualizer(
                isActive: false,
                color: Colors.grey,
                height: 80,
                barCount: 7,
              ),
              const SizedBox(height: 8),
              Text(
                _isChannelBusy ? 'Canal ocupado' : 'Pronto para falar',
                style: TextStyle(
                  color: _isChannelBusy ? Colors.orange : Colors.grey,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
            ],
            const SizedBox(height: 48),

            if (_connectionFailed) ...[
              ElevatedButton.icon(
                onPressed: _retryConnection,
                icon: const Icon(Icons.refresh),
                label: const Text('Tentar novamente'),
              ),
              const SizedBox(height: 16),
              const Text(
                'Não foi possível conectar ao servidor. Verifique se ele está '
                'rodando e na mesma rede.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey, fontSize: 13),
              ),
            ] else
              PttButton(
                isTransmitting: _isTransmitting,
                isChannelBusy: _isChannelBusy,
                onPressed: _startTalking,
                onReleased: _stopTalking,
              ),
            const SizedBox(height: 32),

            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.grey.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  const Text(
                    'Como usar:',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    '• Pressione e segure o botão PTT para falar',
                    style: TextStyle(fontSize: 12),
                  ),
                  const Text(
                    '• Solte para ouvir',
                    style: TextStyle(fontSize: 12),
                  ),
                  const Text(
                    '• Apenas uma pessoa pode falar por vez',
                    style: TextStyle(fontSize: 12),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  String get _currentChannelTitle {
    if (!_webSocketService.isConnected && _connectionError) {
      return '${widget.channelName} (offline)';
    }
    return widget.channelName;
  }

  void _showChannelSettings() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Container(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text(
                'Configurações do Canal',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 16),
              ListTile(
                leading: const Icon(Icons.people),
                title: const Text('Membros do Canal'),
                onTap: () {
                  Navigator.pop(context);
                  _showChannelUsers();
                },
              ),
              ListTile(
                leading: const Icon(Icons.exit_to_app),
                title: const Text('Sair do Canal'),
                onTap: () {
                  Navigator.pop(context);
                  _leaveChannel();
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
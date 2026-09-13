import 'package:flutter/material.dart';
import '../widgets/ptt_button.dart';
import '../widgets/audio_visualizer.dart';
import '../widgets/audio_player.dart';

class VoiceScreen extends StatefulWidget {
  final String channelId;
  final String channelName;

  const VoiceScreen({
    super.key,
    required this.channelId,
    required this.channelName,
  });

  @override
  State<VoiceScreen> createState() => _VoiceScreenState();
}

class _VoiceScreenState extends State<VoiceScreen> {
  bool _isTransmitting = false;
  bool _isReceiving = false;
  bool _isChannelBusy = false;
  String? _currentSpeaker;
  late String _currentChannel;

  @override
  void initState() {
    super.initState();
    _currentChannel = widget.channelName;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_currentChannel),
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
            // Channel info
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.blue.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.radio, color: Colors.blue),
                  const SizedBox(width: 8),
                  Text(
                    _currentChannel,
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 32),

            // Status display
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
                  setState(() {
                    _isReceiving = false;
                  });
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

            // PTT Button
            PttButton(
              isTransmitting: _isTransmitting,
              isChannelBusy: _isChannelBusy,
              onPressed: () {
                setState(() {
                  _isTransmitting = true;
                  _isChannelBusy = true;
                  _currentSpeaker = 'Você';
                });
              },
              onReleased: () {
                setState(() {
                  _isTransmitting = false;
                  _isChannelBusy = false;
                  _currentSpeaker = null;
                });
              },
            ),
            const SizedBox(height: 32),

            // Instructions
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.grey.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Column(
                children: [
                  Text(
                    'Como usar:',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                  SizedBox(height: 8),
                  Text(
                    '• Pressione e segure o botão PTT para falar',
                    style: TextStyle(fontSize: 12),
                  ),
                  Text(
                    '• Solte para ouvir',
                    style: TextStyle(fontSize: 12),
                  ),
                  Text(
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

  void _showChannelSettings() {
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
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
              leading: const Icon(Icons.info),
              title: const Text('Informações do Canal'),
              onTap: () {
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.people),
              title: const Text('Membros do Canal'),
              onTap: () {
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.exit_to_app),
              title: const Text('Sair do Canal'),
              onTap: () {
                Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }
}

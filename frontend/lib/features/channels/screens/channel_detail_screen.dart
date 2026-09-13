import 'package:flutter/material.dart';

import '../models/channel_model.dart';
import '../services/channel_service.dart';
import '../../voice/screens/voice_screen.dart';

class ChannelDetailScreen extends StatefulWidget {
  final String channelId;

  const ChannelDetailScreen({
    super.key,
    required this.channelId,
  });

  @override
  State<ChannelDetailScreen> createState() => _ChannelDetailScreenState();
}

class _ChannelDetailScreenState extends State<ChannelDetailScreen> {
  final ChannelService _channelService = ChannelService();
  ChannelModel? _channel;
  List<String> _users = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadChannelDetails();
  }

  Future<void> _loadChannelDetails() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final channel = await _channelService.getChannel(widget.channelId);
      final users = await _channelService.getChannelUsers(widget.channelId);

      setState(() {
        _channel = channel;
        _users = users;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _error = 'Erro ao carregar detalhes do canal: ${e.toString()}';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_channel?.name ?? 'Canal'),
        actions: [
          if (_channel != null)
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: _loadChannelDetails,
            ),
        ],
      ),
      body: _buildBody(),
      bottomNavigationBar: _channel != null ? _buildBottomBar() : null,
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(),
      );
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red,
            ),
            const SizedBox(height: 16),
            Text(
              _error!,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadChannelDetails,
              child: const Text('Tentar novamente'),
            ),
          ],
        ),
      );
    }

    if (_channel == null) {
      return const Center(
        child: Text('Canal não encontrado'),
      );
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildChannelHeader(),
          const SizedBox(height: 24),
          _buildChannelInfo(),
          const SizedBox(height: 24),
          _buildUsersList(),
        ],
      ),
    );
  }

  Widget _buildChannelHeader() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            CircleAvatar(
              radius: 40,
              backgroundColor: _channel!.isAvailable ? Colors.green : Colors.orange,
              child: const Icon(
                Icons.radio,
                size: 40,
                color: Colors.white,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              _channel!.name,
              style: const TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
              textAlign: TextAlign.center,
            ),
            if (_channel!.description != null &&
                _channel!.description!.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                _channel!.description!,
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.grey[600],
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildChannelInfo() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Informações do Canal',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            _buildInfoRow(
              Icons.people,
              'Usuários',
              '${_channel!.userCount}/${_channel!.maxUsers}',
            ),
            const SizedBox(height: 8),
            _buildInfoRow(
              Icons.radio,
              'Status',
              _channel!.isAvailable ? 'Disponível' : 'Cheio',
            ),
            if (_channel!.distance != null) ...[
              const SizedBox(height: 8),
              _buildInfoRow(
                Icons.location_on,
                'Distância',
                _channel!.distanceFormatted,
              ),
            ],
            if (_channel!.createdAt != null) ...[
              const SizedBox(height: 8),
              _buildInfoRow(
                Icons.calendar_today,
                'Criado em',
                _formatDate(_channel!.createdAt!),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      children: [
        Icon(icon, size: 20, color: Colors.grey),
        const SizedBox(width: 8),
        Text(
          '$label: ',
          style: TextStyle(
            color: Colors.grey[600],
            fontSize: 16,
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }

  Widget _buildUsersList() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Usuários no Canal (${_users.length})',
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            if (_users.isEmpty)
              const Center(
                child: Text(
                  'Nenhum usuário no canal',
                  style: TextStyle(color: Colors.grey),
                ),
              )
            else
              ..._users.map((userId) => ListTile(
                    leading: CircleAvatar(
                      backgroundColor: Colors.blue,
                      child: Text(
                        userId.substring(0, 2).toUpperCase(),
                        style: const TextStyle(color: Colors.white),
                      ),
                    ),
                    title: Text(userId),
                    dense: true,
                  )),
          ],
        ),
      ),
    );
  }

  Widget _buildBottomBar() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          Expanded(
            child: ElevatedButton.icon(
              onPressed: _channel!.isAvailable ? _joinChannel : null,
              icon: const Icon(Icons.login),
              label: const Text('Entrar no Canal'),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.green,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: ElevatedButton.icon(
              onPressed: _enterVoiceChat,
              icon: const Icon(Icons.mic),
              label: const Text('Falar'),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blue,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }

  Future<void> _joinChannel() async {
    try {
      await _channelService.joinChannel(widget.channelId);
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Entrou no canal ${_channel!.name}'),
            backgroundColor: Colors.green,
          ),
        );
        _loadChannelDetails();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Erro ao entrar no canal: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _enterVoiceChat() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => VoiceScreen(
          channelId: widget.channelId,
          channelName: _channel!.name,
        ),
      ),
    );
  }
}

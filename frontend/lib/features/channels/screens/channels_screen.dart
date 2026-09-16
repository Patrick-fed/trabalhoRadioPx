import 'package:flutter/material.dart';

import '../models/channel_model.dart';
import '../services/channel_service.dart';
import '../../location/services/location_service.dart';

class ChannelsScreen extends StatefulWidget {
  final void Function(String channelId, String channelName)? onChannelSelected;

  const ChannelsScreen({super.key, this.onChannelSelected});

  @override
  State<ChannelsScreen> createState() => _ChannelsScreenState();
}

class _ChannelsScreenState extends State<ChannelsScreen> {
  final ChannelService _channelService = ChannelService();
  final LocationService _locationService = LocationService();
  
  List<ChannelModel> _channels = [];
  bool _isLoading = true;
  bool _hasLocationPermission = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _initializeLocation();
  }

  Future<void> _initializeLocation() async {
    _hasLocationPermission = await _locationService.checkPermission();
    
    if (_hasLocationPermission) {
      await _loadNearbyChannels();
    } else {
      setState(() {
        _isLoading = false;
        _error = 'Permissão de localização necessária para encontrar canais próximos';
      });
    }
  }

  Future<void> _loadNearbyChannels() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final location = await _locationService.getCurrentLocation();
      if (location == null) {
        setState(() {
          _isLoading = false;
          _error = 'Não foi possível obter a localização atual';
        });
        return;
      }

      final channels = await _channelService.getNearbyChannels(
        latitude: location.latitude,
        longitude: location.longitude,
      );

      setState(() {
        _channels = channels;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _error = 'Erro ao carregar canais: ${e.toString()}';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Canais Próximos'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadNearbyChannels,
          ),
        ],
      ),
      body: _buildBody(),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreateChannelDialog,
        child: const Icon(Icons.add),
      ),
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
              onPressed: _loadNearbyChannels,
              child: const Text('Tentar novamente'),
            ),
          ],
        ),
      );
    }

    if (_channels.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.radio,
              size: 64,
              color: Colors.grey,
            ),
            const SizedBox(height: 16),
            const Text(
              'Nenhum canal próximo encontrado',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 8),
            const Text(
              'Mova-se para encontrar canais',
              style: TextStyle(color: Colors.grey),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadNearbyChannels,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _channels.length,
        itemBuilder: (context, index) {
          final channel = _channels[index];
          return _buildChannelCard(channel);
        },
      ),
    );
  }

  Widget _buildChannelCard(ChannelModel channel) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: ListTile(
        contentPadding: const EdgeInsets.all(16),
        leading: CircleAvatar(
          backgroundColor: channel.isAvailable ? Colors.green : Colors.orange,
          radius: 24,
          child: Icon(
            Icons.radio,
            color: Colors.white,
          ),
        ),
        title: Text(
          channel.name,
          style: const TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Row(
              children: [
                const Icon(Icons.people, size: 16, color: Colors.grey),
                const SizedBox(width: 4),
                Text(
                  '${channel.userCount}/${channel.maxUsers} usuários',
                  style: TextStyle(color: Colors.grey[600]),
                ),
              ],
            ),
            if (channel.distance != null) ...[
              const SizedBox(height: 4),
              Row(
                children: [
                  const Icon(Icons.location_on, size: 16, color: Colors.grey),
                  const SizedBox(width: 4),
                  Text(
                    _formatDistance(channel.distance!),
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                ],
              ),
            ],
          ],
        ),
        trailing: channel.isAvailable
            ? const Icon(Icons.chevron_right, color: Colors.green)
            : const Icon(Icons.lock, color: Colors.orange),
        onTap: () => _joinChannel(channel),
      ),
    );
  }

  String _formatDistance(double distanceKm) {
    if (distanceKm < 1) {
      return '${(distanceKm * 1000).round()}m';
    }
    return '${distanceKm.toStringAsFixed(1)}km';
  }

  Future<void> _joinChannel(ChannelModel channel) async {
    if (!channel.isAvailable) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Canal está cheio'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    try {
      await _channelService.joinChannel(channel.id);
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Entrou no canal ${channel.name}'),
            backgroundColor: Colors.green,
          ),
        );

        widget.onChannelSelected?.call(channel.id, channel.name);
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

  void _showCreateChannelDialog() {
    final nameController = TextEditingController();
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Criar Canal'),
        content: TextField(
          controller: nameController,
          decoration: const InputDecoration(
            hintText: 'Nome do canal',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancelar'),
          ),
          ElevatedButton(
            onPressed: () async {
              if (nameController.text.isNotEmpty) {
                await _createChannel(nameController.text);
                Navigator.pop(context);
              }
            },
            child: const Text('Criar'),
          ),
        ],
      ),
    );
  }

  Future<void> _createChannel(String name) async {
    try {
      final location = await _locationService.getCurrentLocation();
      if (location == null) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Localização necessária para criar canal'),
              backgroundColor: Colors.orange,
            ),
          );
        }
        return;
      }

      await _channelService.createChannel(
        name: name,
        latitude: location.latitude,
        longitude: location.longitude,
      );

      await _loadNearbyChannels();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Erro ao criar canal: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }
}

import 'dart:async';

import 'package:flutter/material.dart';

import '../../../core/api_client.dart';
import '../models/channel_model.dart';
import '../services/channel_service.dart';
import '../../location/models/location_model.dart';
import '../../location/services/location_service.dart';

class ChannelsScreen extends StatefulWidget {
  final void Function(String channelId, String channelName)? onChannelSelected;

  const ChannelsScreen({super.key, this.onChannelSelected});

  @override
  State<ChannelsScreen> createState() => _ChannelsScreenState();
}

class _ChannelsScreenState extends State<ChannelsScreen>
    with SingleTickerProviderStateMixin, WidgetsBindingObserver {
  final ChannelService _channelService = ChannelService();
  final LocationService _locationService = LocationService();

  late final TabController _tabController;

  static const Duration _autoRefreshInterval = Duration(seconds: 10);
  static const Duration _locationTtl = Duration(minutes: 2);

  List<ChannelModel> _myChannels = [];
  List<ChannelModel> _nearbyChannels = [];
  bool _isLoading = true;
  String? _error;

  Timer? _refreshTimer;
  LocationModel? _cachedLocation;
  DateTime? _cachedLocationAt;
  bool _isRefreshing = false;

  static const int _myChannelsTab = 0;
  static const int _nearbyChannelsTab = 1;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    WidgetsBinding.instance.addObserver(this);
    _startAutoRefresh();
    _initialize();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _stopAutoRefresh();
    _tabController.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.resumed:
        _startAutoRefresh();
        _silentRefresh();
      case AppLifecycleState.inactive:
      case AppLifecycleState.paused:
      case AppLifecycleState.hidden:
      case AppLifecycleState.detached:
        _stopAutoRefresh();
    }
  }

  void _startAutoRefresh() {
    _refreshTimer ??= Timer.periodic(
      _autoRefreshInterval,
      (_) => _silentRefresh(),
    );
  }

  void _stopAutoRefresh() {
    _refreshTimer?.cancel();
    _refreshTimer = null;
  }

  Future<void> _initialize() async {
    final hasPermission = await _locationService.checkPermission();

    if (!mounted) return;

    if (hasPermission) {
      await _loadAll();
    } else {
      setState(() {
        _isLoading = false;
        _error = 'Permissão de localização necessária para encontrar canais próximos';
      });
    }
  }

  Future<void> _loadAll() => _refresh(showLoading: true);

  Future<void> _silentRefresh() => _refresh(showLoading: false);

  Future<void> _refresh({required bool showLoading}) async {
    if (_isRefreshing) return;
    _isRefreshing = true;

    if (showLoading) {
      setState(() {
        _isLoading = true;
        _error = null;
      });
    }

    try {
      final location = await _resolveLocation(force: showLoading);
      if (location == null) {
        if (!mounted || !showLoading) return;
        setState(() {
          _isLoading = false;
          _error = 'Não foi possível obter a localização atual';
        });
        return;
      }

      final results = await Future.wait([
        _channelService.getUserChannels(),
        _channelService.getNearbyChannels(
          latitude: location.latitude,
          longitude: location.longitude,
        ),
      ]);

      if (!mounted) return;

      final myChannels = results[0];
      final nearbyChannels = results[1];

      final unchanged = _sameChannels(_myChannels, myChannels) &&
          _sameChannels(_nearbyChannels, nearbyChannels);
      if (unchanged && !showLoading) return;

      setState(() {
        _myChannels = myChannels;
        _nearbyChannels = nearbyChannels;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted || !showLoading) return;
      setState(() {
        _isLoading = false;
        _error = 'Erro ao carregar canais: ${friendlyError(e)}';
      });
    } finally {
      _isRefreshing = false;
    }
  }

  Future<LocationModel?> _resolveLocation({bool force = false}) async {
    final cached = _cachedLocation;
    final cachedAt = _cachedLocationAt;
    final isFresh = !force &&
        cached != null &&
        cachedAt != null &&
        DateTime.now().difference(cachedAt) < _locationTtl;

    if (isFresh) return cached;

    final location = await _locationService.getCurrentLocation();
    if (location != null) {
      _cachedLocation = location;
      _cachedLocationAt = DateTime.now();
    }

    return location;
  }

  bool _sameChannels(List<ChannelModel> current, List<ChannelModel> incoming) {
    if (current.length != incoming.length) return false;

    for (var i = 0; i < current.length; i++) {
      final a = current[i];
      final b = incoming[i];
      if (a.id != b.id ||
          a.userCount != b.userCount ||
          a.isAvailable != b.isAvailable ||
          a.updatedAt != b.updatedAt ||
          a.distance != b.distance) {
        return false;
      }
    }

    return true;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Canais'),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Meus canais', icon: Icon(Icons.person_pin_circle)),
            Tab(text: 'Perto de mim', icon: Icon(Icons.near_me)),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadAll,
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
      return _buildMessageState(
        icon: Icons.error_outline,
        iconColor: Colors.red,
        message: _error!,
        actionLabel: 'Tentar novamente',
        onAction: _loadAll,
      );
    }

    return TabBarView(
      controller: _tabController,
      children: [
        _buildChannelList(
          channels: _myChannels,
          emptyIcon: Icons.person_pin_circle,
          emptyMessage: 'Você ainda não entrou em nenhum canal',
          emptyHint: 'Entre em um canal próximo para falar',
          emptyActionLabel: 'Ver canais perto de mim',
          onEmptyAction: () => _goToTab(_nearbyChannelsTab),
          onChannelTap: _openVoice,
          trailingBuilder: (channel) => IconButton(
            tooltip: 'Sair do canal',
            icon: const Icon(Icons.exit_to_app),
            color: Colors.orange,
            onPressed: () => _leaveChannel(channel),
          ),
        ),
        _buildChannelList(
          channels: _nearbyChannels,
          emptyIcon: Icons.near_me,
          emptyMessage: 'Nenhum canal próximo encontrado',
          emptyHint: 'Mova-se ou crie um canal para começar',
          emptyActionLabel: 'Criar canal',
          onEmptyAction: _showCreateChannelDialog,
          onChannelTap: _joinChannel,
        ),
      ],
    );
  }

  Widget _buildChannelList({
    required List<ChannelModel> channels,
    required IconData emptyIcon,
    required String emptyMessage,
    required String emptyHint,
    required String emptyActionLabel,
    required VoidCallback onEmptyAction,
    required void Function(ChannelModel channel) onChannelTap,
    Widget Function(ChannelModel channel)? trailingBuilder,
  }) {
    if (channels.isEmpty) {
      return _buildMessageState(
        icon: emptyIcon,
        iconColor: Colors.grey,
        message: emptyMessage,
        hint: emptyHint,
        actionLabel: emptyActionLabel,
        onAction: onEmptyAction,
      );
    }

    return RefreshIndicator(
      onRefresh: _loadAll,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: channels.length,
        itemBuilder: (context, index) {
          return _buildChannelCard(
            channels[index],
            onTap: () => onChannelTap(channels[index]),
            trailingBuilder: trailingBuilder,
          );
        },
      ),
    );
  }

  Widget _buildMessageState({
    required IconData icon,
    required Color iconColor,
    required String message,
    String? hint,
    required String actionLabel,
    required VoidCallback onAction,
  }) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 64, color: iconColor),
            const SizedBox(height: 16),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 16),
            ),
            if (hint != null) ...[
              const SizedBox(height: 8),
              Text(
                hint,
                textAlign: TextAlign.center,
                style: const TextStyle(color: Colors.grey),
              ),
            ],
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: onAction,
              child: Text(actionLabel),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildChannelCard(
    ChannelModel channel, {
    required VoidCallback onTap,
    Widget Function(ChannelModel channel)? trailingBuilder,
  }) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: ListTile(
        contentPadding: const EdgeInsets.all(16),
        leading: CircleAvatar(
          backgroundColor: channel.isAvailable ? Colors.green : Colors.orange,
          radius: 24,
          child: const Icon(Icons.radio, color: Colors.white),
        ),
        title: Text(
          channel.name,
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18),
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
                    channel.distanceFormatted,
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                ],
              ),
            ],
          ],
        ),
        trailing: trailingBuilder?.call(channel) ??
            (channel.isAvailable
                ? const Icon(Icons.chevron_right, color: Colors.green)
                : const Icon(Icons.lock, color: Colors.orange)),
        onTap: onTap,
      ),
    );
  }

  void _goToTab(int index) {
    _tabController.animateTo(index);
  }

  void _openVoice(ChannelModel channel) {
    widget.onChannelSelected?.call(channel.id, channel.name);
  }

  Future<void> _joinChannel(ChannelModel channel) async {
    if (!channel.isAvailable) {
      _showSnack('Canal está cheio', Colors.orange);
      return;
    }

    try {
      await _channelService.joinChannel(channel.id);
    } catch (e) {
      if (!mounted) return;
      _showSnack('Erro ao entrar no canal: ${friendlyError(e)}', Colors.red);
      return;
    }

    if (!mounted) return;
    _showSnack('Entrou no canal ${channel.name}', Colors.green);

    _goToTab(_myChannelsTab);
    await _loadAll();
  }

  Future<void> _leaveChannel(ChannelModel channel) async {
    try {
      await _channelService.leaveChannel(channel.id);
    } catch (e) {
      if (!mounted) return;
      _showSnack('Erro ao sair do canal: ${friendlyError(e)}', Colors.red);
      return;
    }

    if (!mounted) return;
    _showSnack('Saiu do canal ${channel.name}', Colors.orange);

    await _loadAll();
  }

  void _showSnack(String message, Color backgroundColor) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message), backgroundColor: backgroundColor));
  }

  Future<void> _showCreateChannelDialog() async {
    final nameController = TextEditingController();

    await showDialog<void>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Criar Canal'),
        content: TextField(
          controller: nameController,
          decoration: const InputDecoration(hintText: 'Nome do canal'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Cancelar'),
          ),
          ElevatedButton(
            onPressed: () async {
              final name = nameController.text.trim();
              if (name.isEmpty) return;
              Navigator.pop(dialogContext);
              await _createChannel(name);
            },
            child: const Text('Criar'),
          ),
        ],
      ),
    );

    nameController.dispose();
  }

  Future<void> _createChannel(String name) async {
    final location = await _locationService.getCurrentLocation();
    if (location == null) {
      if (!mounted) return;
      _showSnack('Localização necessária para criar canal', Colors.orange);
      return;
    }

    try {
      await _channelService.createChannel(
        name: name,
        latitude: location.latitude,
        longitude: location.longitude,
      );
    } catch (e) {
      if (!mounted) return;
      _showSnack('Erro ao criar canal: ${friendlyError(e)}', Colors.red);
      return;
    }

    if (!mounted) return;
    _showSnack('Canal $name criado', Colors.green);

    _goToTab(_myChannelsTab);
    await _loadAll();
  }
}

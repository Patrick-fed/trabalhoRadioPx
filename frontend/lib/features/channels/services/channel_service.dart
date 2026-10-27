import 'dart:convert';

import '../models/channel_model.dart';
import '../../auth/services/auth_service.dart';
import '../../../core/api_client.dart';
import '../../../core/config.dart';

class ChannelService {
  String get baseUrl => AppConfig.apiBaseUrl;
  final AuthService _authService = AuthService();

  Future<List<ChannelModel>> getNearbyChannels({
    required double latitude,
    required double longitude,
    double radiusKm = 10.0,
  }) async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiGet(
      '$baseUrl/api/v1/channels/nearby?lat=$latitude&lng=$longitude&radius=$radiusKm',
      headers: headers,
    );

    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((json) => ChannelModel.fromJson(json)).toList();
    } else {
      throw Exception('Falha ao carregar canais próximos');
    }
  }

  Future<ChannelModel> getChannel(String channelId) async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiGet(
      '$baseUrl/api/v1/channels/$channelId',
      headers: headers,
    );

    if (response.statusCode == 200) {
      return ChannelModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Falha ao carregar o canal');
    }
  }

  Future<ChannelModel> createChannel({
    required String name,
    required double latitude,
    required double longitude,
  }) async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiPost(
      '$baseUrl/api/v1/channels',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: jsonEncode({
        'name': name,
        'latitude': latitude,
        'longitude': longitude,
      }),
    );

    if (response.statusCode == 201) {
      return ChannelModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Falha ao criar o canal');
    }
  }

  Future<void> joinChannel(String channelId) async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiPost(
      '$baseUrl/api/v1/channels/$channelId/join',
      headers: headers,
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Falha ao entrar no canal (${response.statusCode}): ${response.body}',
      );
    }
  }

  Future<void> leaveChannel(String channelId) async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiPost(
      '$baseUrl/api/v1/channels/$channelId/leave',
      headers: headers,
    );

    if (response.statusCode != 200) {
      throw Exception(
        'Falha ao sair do canal (${response.statusCode}): ${response.body}',
      );
    }
  }

  Future<List<ChannelModel>> getUserChannels() async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiGet(
      '$baseUrl/api/v1/channels',
      headers: headers,
    );

    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((json) => ChannelModel.fromJson(json)).toList();
    } else {
      throw Exception('Falha ao carregar seus canais');
    }
  }

  Future<List<String>> getChannelUsers(String channelId) async {
    final headers = await _authService.getAuthHeaders();
    final response = await apiGet(
      '$baseUrl/api/v1/channels/$channelId/users',
      headers: headers,
    );

    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((e) => e.toString()).toList();
    } else {
      throw Exception('Falha ao carregar os membros do canal');
    }
  }
}

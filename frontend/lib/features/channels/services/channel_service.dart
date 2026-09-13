import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/channel_model.dart';

class ChannelService {
  final String baseUrl = 'http://localhost:8080';

  Future<List<ChannelModel>> getNearbyChannels({
    required double latitude,
    required double longitude,
    double radiusKm = 10.0,
  }) async {
    final response = await http.get(
      Uri.parse(
        '$baseUrl/api/v1/channels/nearby?lat=$latitude&lng=$longitude&radius=$radiusKm',
      ),
    );

    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((json) => ChannelModel.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load nearby channels');
    }
  }

  Future<ChannelModel> getChannel(String channelId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/v1/channels/$channelId'),
    );

    if (response.statusCode == 200) {
      return ChannelModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to load channel');
    }
  }

  Future<ChannelModel> createChannel({
    required String name,
    required double latitude,
    required double longitude,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/channels'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'name': name,
        'latitude': latitude,
        'longitude': longitude,
      }),
    );

    if (response.statusCode == 201) {
      return ChannelModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to create channel');
    }
  }

  Future<void> joinChannel(String channelId) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/channels/$channelId/join'),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to join channel');
    }
  }

  Future<void> leaveChannel(String channelId) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/channels/$channelId/leave'),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to leave channel');
    }
  }

  Future<List<ChannelModel>> getUserChannels() async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/v1/channels/user'),
    );

    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((json) => ChannelModel.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load user channels');
    }
  }
}

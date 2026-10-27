import 'dart:convert';

import '../../../core/api_client.dart';
import '../../../core/config.dart';
import '../models/profile_model.dart';
import '../../auth/services/auth_service.dart';

class ProfileService {
  String get baseUrl => AppConfig.apiBaseUrl;
  final AuthService _authService = AuthService();

  Future<ProfileModel> getProfile() async {
    final headers = await _authService.getAuthHeaders();

    final response = await apiGet(
      '$baseUrl/api/v1/user/profile',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
    );

    if (response.statusCode == 200) {
      return ProfileModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Falha ao carregar o perfil');
    }
  }

  Future<ProfileModel> updateProfile({
    String? name,
    String? email,
    String? avatar,
  }) async {
    final headers = await _authService.getAuthHeaders();

    final response = await apiPut(
      '$baseUrl/api/v1/user/profile',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: jsonEncode({
        if (name != null) 'name': name,
        if (email != null) 'email': email,
        if (avatar != null) 'avatar': avatar,
      }),
    );

    if (response.statusCode == 200) {
      return ProfileModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Falha ao atualizar o perfil');
    }
  }

  Future<void> changePassword({
    required String oldPassword,
    required String newPassword,
  }) async {
    final headers = await _authService.getAuthHeaders();

    final response = await apiPost(
      '$baseUrl/api/v1/user/change-password',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: jsonEncode({
        'old_password': oldPassword,
        'new_password': newPassword,
      }),
    );

    if (response.statusCode != 200) {
      throw Exception('Falha ao alterar a senha');
    }
  }

  Future<void> deleteAccount() async {
    final headers = await _authService.getAuthHeaders();

    final response = await apiDelete(
      '$baseUrl/api/v1/user/profile',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
    );

    if (response.statusCode != 200) {
      throw Exception('Falha ao excluir a conta');
    }
  }
}

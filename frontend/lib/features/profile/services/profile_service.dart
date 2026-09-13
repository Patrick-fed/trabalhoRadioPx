import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/profile_model.dart';
import '../../auth/services/auth_service.dart';

class ProfileService {
  final String baseUrl = 'http://localhost:8080';
  final AuthService _authService = AuthService();

  Future<ProfileModel> getProfile() async {
    final headers = await _authService.getAuthHeaders();
    
    final response = await http.get(
      Uri.parse('$baseUrl/api/v1/user/profile'),
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
    );

    if (response.statusCode == 200) {
      return ProfileModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to load profile');
    }
  }

  Future<ProfileModel> updateProfile({
    String? name,
    String? email,
    String? avatar,
  }) async {
    final headers = await _authService.getAuthHeaders();
    
    final response = await http.put(
      Uri.parse('$baseUrl/api/v1/user/profile'),
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
      throw Exception('Failed to update profile');
    }
  }

  Future<void> changePassword({
    required String oldPassword,
    required String newPassword,
  }) async {
    final headers = await _authService.getAuthHeaders();
    
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/user/change-password'),
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
      throw Exception('Failed to change password');
    }
  }

  Future<void> deleteAccount() async {
    final headers = await _authService.getAuthHeaders();
    
    final response = await http.delete(
      Uri.parse('$baseUrl/api/v1/user/profile'),
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to delete account');
    }
  }
}

import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/config.dart';
import '../models/user_model.dart';

class AuthService {
  final String baseUrl = AppConfig.apiBaseUrl;
  static const String _tokenKey = 'auth_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _userKey = 'user_data';

  Future<UserModel> login({
    required String email,
    required String password,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'email': email,
        'password': password,
      }),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final token = data['token'] as String;
      final user = UserModel.fromJson(data['user']);
      
      await _saveAuthData(token, data['refresh_token'], user);
      
      return user;
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Login failed');
    }
  }

  Future<UserModel> register({
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/auth/register'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'name': name,
        'email': email,
        'password': password,
      }),
    );

    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      final token = data['token'] as String;
      final user = UserModel.fromJson(data['user']);
      
      await _saveAuthData(token, data['refresh_token'], user);
      
      return user;
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Registration failed');
    }
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    await prefs.remove(_refreshTokenKey);
    await prefs.remove(_userKey);
  }

  Future<UserModel?> getCurrentUser() async {
    final prefs = await SharedPreferences.getInstance();
    final userJson = prefs.getString(_userKey);
    
    if (userJson == null) return null;
    
    return UserModel.fromJson(jsonDecode(userJson));
  }

  Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }

  Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null;
  }

  Future<UserModel?> refreshToken() async {
    final prefs = await SharedPreferences.getInstance();
    final refreshToken = prefs.getString(_refreshTokenKey);
    
    if (refreshToken == null) return null;

    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/auth/refresh'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'refresh_token': refreshToken,
      }),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final token = data['token'] as String;
      final user = UserModel.fromJson(data['user']);
      
      await _saveAuthData(token, data['refresh_token'], user);
      
      return user;
    } else {
      await logout();
      return null;
    }
  }

  Future<void> _saveAuthData(String token, String? refreshToken, UserModel user) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
    if (refreshToken != null) {
      await prefs.setString(_refreshTokenKey, refreshToken);
    }
    await prefs.setString(_userKey, jsonEncode(user.toJson()));
  }

  Future<Map<String, String>> getAuthHeaders() async {
    final token = await getToken();
    if (token == null) return {};
    
    return {
      'Authorization': 'Bearer $token',
    };
  }
}

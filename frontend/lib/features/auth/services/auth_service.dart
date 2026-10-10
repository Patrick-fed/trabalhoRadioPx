import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/api_client.dart';
import '../../../core/config.dart';
import '../models/user_model.dart';
import 'token_validity.dart';

class AuthService {
  final String baseUrl = AppConfig.apiBaseUrl;
  static const String _tokenKey = 'auth_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _userKey = 'user_data';

  static final ValueNotifier<bool> sessionState = ValueNotifier<bool>(false);

  bool _refreshingSession = false;

  Future<UserModel> login({
    required String email,
    required String password,
  }) async {
    final response = await apiPost(
      '$baseUrl/api/v1/auth/login',
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
      throw Exception(_readError(response));
    }
  }

  Future<UserModel> register({
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await apiPost(
      '$baseUrl/api/v1/auth/register',
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'name': name,
        'email': email,
        'password': password,
      }),
    );

    if (response.statusCode == 200 || response.statusCode == 201) {
      final data = jsonDecode(response.body);
      final token = data['token'] as String;
      final user = UserModel.fromJson(data['user']);
      
      await _saveAuthData(token, data['refresh_token'], user);
      
      return user;
    } else {
      throw Exception(_readError(response));
    }
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    await prefs.remove(_refreshTokenKey);
    await prefs.remove(_userKey);
    sessionState.value = false;
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
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString(_tokenKey);

    if (token == null) return false;
    if (!tokenIsExpired(token)) return true;

    // Access token expired: try the silent renewal before giving up.
    if (_refreshingSession) return false;
    final savedRefreshToken = prefs.getString(_refreshTokenKey);
    if (savedRefreshToken == null) return false;

    _refreshingSession = true;
    try {
      final user = await refreshToken();
      return user != null;
    } catch (_) {
      return false;
    } finally {
      _refreshingSession = false;
    }
  }

  Future<UserModel?> refreshToken() async {
    final prefs = await SharedPreferences.getInstance();
    final refreshToken = prefs.getString(_refreshTokenKey);
    
    if (refreshToken == null) return null;

    final response = await apiPost(
      '$baseUrl/api/v1/auth/refresh',
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
    sessionState.value = true;
  }

  String _readError(http.Response response) {
    try {
      final error = jsonDecode(response.body);
      if (error is Map && error['error'] is String) {
        return error['error'] as String;
      }
    } catch (_) {}
    return 'Falha na requisição (${response.statusCode})';
  }

  Future<Map<String, String>> getAuthHeaders() async {
    final prefs = await SharedPreferences.getInstance();
    var token = prefs.getString(_tokenKey);
    if (token == null) return {};

    if (tokenIsExpired(token)) {
      // Expired access token on a real API call: try to renew once. On failure
      // refreshToken() logs out, so no 401 wall of errors reaches the UI.
      final user = await refreshToken();
      if (user == null) return {};
      token = prefs.getString(_tokenKey);
    }

    if (token == null || token.isEmpty) return {};
    return {
      'Authorization': 'Bearer $token',
    };
  }
}

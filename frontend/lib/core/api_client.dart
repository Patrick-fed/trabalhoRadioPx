import 'dart:async';
import 'dart:io';

import 'package:http/http.dart' as http;

const Duration apiTimeout = Duration(seconds: 10);

const String _serverUnreachable =
    'Servidor não respondeu. Verifique a conexão com a rede.';

const String _serverUnconnectable =
    'Não foi possível conectar ao servidor. Verifique se ele está rodando e na mesma rede.';

Future<http.Response> apiGet(
  String url, {
  Map<String, String>? headers,
}) {
  return _withTimeout(() => http.get(Uri.parse(url), headers: headers));
}

Future<http.Response> apiPost(
  String url, {
  Object? body,
  Map<String, String>? headers,
}) {
  return _withTimeout(() => http.post(Uri.parse(url), body: body, headers: headers));
}

Future<http.Response> apiPut(
  String url, {
  Object? body,
  Map<String, String>? headers,
}) {
  return _withTimeout(() => http.put(Uri.parse(url), body: body, headers: headers));
}

Future<http.Response> apiDelete(
  String url, {
  Map<String, String>? headers,
}) {
  return _withTimeout(() => http.delete(Uri.parse(url), headers: headers));
}

Future<http.Response> _withTimeout(
  Future<http.Response> Function() request,
) async {
  try {
    return await request().timeout(apiTimeout);
  } on TimeoutException {
    throw Exception(_serverUnreachable);
  } on SocketException {
    throw Exception(_serverUnconnectable);
  } on http.ClientException catch (e) {
    throw Exception('Falha de conexão com o servidor: ${e.message}');
  }
}

String friendlyError(Object error) {
  final text = error.toString();
  const prefix = 'Exception: ';
  if (text.startsWith(prefix)) {
    return text.substring(prefix.length);
  }
  return text;
}

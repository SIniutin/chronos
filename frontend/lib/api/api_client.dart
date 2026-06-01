import 'dart:convert';
import 'dart:typed_data';

import 'package:http/http.dart' as http;
import 'package:http_parser/http_parser.dart';

import 'api_config.dart';

class ApiException implements Exception {
  final int statusCode;
  final String message;

  const ApiException(this.statusCode, this.message);

  @override
  String toString() => message;
}

typedef RefreshCallback = Future<bool> Function();
typedef AccessTokenReader = String? Function();

class ApiClient {
  final http.Client _client;
  final String baseUrl;
  AccessTokenReader? accessTokenReader;
  RefreshCallback? onUnauthorized;

  ApiClient({
    http.Client? client,
    this.baseUrl = ApiConfig.baseUrl,
    this.accessTokenReader,
    this.onUnauthorized,
  }) : _client = client ?? http.Client();

  Future<dynamic> get(String path, {bool auth = false}) {
    return _send('GET', path, auth: auth);
  }

  Future<dynamic> post(String path, {Object? body, bool auth = false}) {
    return _send('POST', path, body: body, auth: auth);
  }

  Future<dynamic> patch(String path, {Object? body, bool auth = false}) {
    return _send('PATCH', path, body: body, auth: auth);
  }

  Future<dynamic> uploadBytes(
    String path, {
    required String fieldName,
    required String filename,
    required Uint8List bytes,
    required String contentType,
    bool auth = false,
    bool retried = false,
  }) async {
    final uri = Uri.parse('$baseUrl$path');
    final request = http.MultipartRequest('POST', uri)
      ..headers['Accept'] = 'application/json'
      ..files.add(http.MultipartFile.fromBytes(
        fieldName,
        bytes,
        filename: filename,
        contentType: _mediaType(contentType),
      ));
    final token = accessTokenReader?.call();
    if (auth && token != null && token.isNotEmpty) {
      request.headers['Authorization'] = 'Bearer $token';
    }

    final streamed = await _client.send(request);
    final response = await http.Response.fromStream(streamed);
    if (response.statusCode == 401 && auth && !retried && onUnauthorized != null) {
      final refreshed = await onUnauthorized!.call();
      if (refreshed) {
        return uploadBytes(
          path,
          fieldName: fieldName,
          filename: filename,
          bytes: bytes,
          contentType: contentType,
          auth: auth,
          retried: true,
        );
      }
    }
    return _decodeResponse(response);
  }

  Future<dynamic> _send(
    String method,
    String path, {
    Object? body,
    bool auth = false,
    bool retried = false,
  }) async {
    final uri = Uri.parse('$baseUrl$path');
    final headers = <String, String>{
      'Accept': 'application/json',
      if (body != null) 'Content-Type': 'application/json',
    };
    final token = accessTokenReader?.call();
    if (auth && token != null && token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }

    final encodedBody = body == null ? null : jsonEncode(body);
    final response = switch (method) {
      'GET' => await _client.get(uri, headers: headers),
      'POST' => await _client.post(uri, headers: headers, body: encodedBody),
      'PATCH' => await _client.patch(uri, headers: headers, body: encodedBody),
      _ => throw ArgumentError('unsupported method $method'),
    };

    if (response.statusCode == 401 && auth && !retried && onUnauthorized != null) {
      final refreshed = await onUnauthorized!.call();
      if (refreshed) {
        return _send(method, path, body: body, auth: auth, retried: true);
      }
    }

    return _decodeResponse(response);
  }

  dynamic _decodeResponse(http.Response response) {
    if (response.statusCode == 204) {
      return null;
    }
    final decoded = response.body.isEmpty ? null : jsonDecode(response.body);
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return decoded;
    }

    var message = 'Ошибка сервера';
    if (decoded is Map && decoded['error'] is String) {
      message = decoded['error'] as String;
    }
    throw ApiException(response.statusCode, message);
  }
}

MediaType _mediaType(String raw) {
  final parts = raw.split('/');
  if (parts.length == 2) {
    return MediaType(parts[0], parts[1]);
  }
  return MediaType('application', 'octet-stream');
}

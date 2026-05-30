import 'api_client.dart';

class AppUser {
  final String? id;
  final String email;
  final String login;
  final String role;
  final String createdAt;

  const AppUser({
    this.id,
    required this.email,
    required this.login,
    required this.role,
    required this.createdAt,
  });

  bool get canOpenPanel =>
      role == 'content_editor' || role == 'content_reviewer' || role == 'admin';

  bool get canEditContent => role == 'content_editor' || role == 'admin';
  bool get canReviewContent => role == 'content_reviewer' || role == 'admin';
  bool get isAdmin => role == 'admin';

  factory AppUser.fromJson(Map<String, dynamic> json) {
    return AppUser(
      id: json['id'] as String?,
      email: json['email'] as String? ?? '',
      login: json['login'] as String? ?? '',
      role: json['role'] as String? ?? 'student',
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}

class TokenPair {
  final String accessToken;
  final String refreshToken;

  const TokenPair({
    required this.accessToken,
    required this.refreshToken,
  });

  factory TokenPair.fromJson(Map<String, dynamic> json) {
    return TokenPair(
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
    );
  }
}

class AuthApi {
  final ApiClient _client;

  const AuthApi(this._client);

  Future<AppUser> register({
    required String email,
    required String login,
    required String password,
  }) async {
    final json = await _client.post('/auth/register', body: {
      'email': email,
      'login': login,
      'password': password,
    });
    return AppUser.fromJson(json as Map<String, dynamic>);
  }

  Future<TokenPair> login({
    required String identity,
    required String password,
  }) async {
    final json = await _client.post('/auth/login', body: {
      'identity': identity,
      'password': password,
    });
    return TokenPair.fromJson(json as Map<String, dynamic>);
  }

  Future<TokenPair> refresh(String refreshToken) async {
    final json = await _client.post('/auth/refresh', body: {
      'refresh_token': refreshToken,
    });
    return TokenPair.fromJson(json as Map<String, dynamic>);
  }

  Future<void> logout(String refreshToken) async {
    await _client.post('/auth/logout', body: {
      'refresh_token': refreshToken,
    }, auth: true);
  }

  Future<AppUser> me() async {
    final json = await _client.get('/users/me', auth: true);
    return AppUser.fromJson(json as Map<String, dynamic>);
  }

  Future<AppUser> findUser(String identity) async {
    final encoded = Uri.encodeQueryComponent(identity);
    final json = await _client.get('/admin/users/lookup?identity=$encoded', auth: true);
    return AppUser.fromJson(json as Map<String, dynamic>);
  }

  Future<AppUser> changeRole(String userId, String role) async {
    final json = await _client.patch('/admin/users/$userId/role', body: {
      'role': role,
    }, auth: true);
    final data = Map<String, dynamic>.from(json as Map<String, dynamic>);
    data['id'] ??= userId;
    return AppUser.fromJson(data);
  }
}

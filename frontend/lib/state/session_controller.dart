import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../api/api_client.dart';
import '../api/auth_api.dart';
import '../theme/app_theme.dart';

class SessionController extends ChangeNotifier {
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';

  final ApiClient client;
  final FlutterSecureStorage storage;
  late final AuthApi authApi;

  String? accessToken;
  String? refreshToken;
  AppUser? currentUser;
  bool isBusy = false;
  bool isRestoring = false;
  ThemeMode themeMode = ThemeMode.dark;

  SessionController({
    ApiClient? apiClient,
    FlutterSecureStorage? secureStorage,
  })  : client = apiClient ?? ApiClient(),
        storage = secureStorage ?? const FlutterSecureStorage() {
    authApi = AuthApi(client);
    client.accessTokenReader = () => accessToken;
    client.onUnauthorized = refresh;
  }

  bool get isAuthenticated => accessToken != null && currentUser != null;
  bool get isLightTheme => themeMode == ThemeMode.light;

  void setLightTheme(bool value) {
    themeMode = value ? ThemeMode.light : ThemeMode.dark;
    AppTheme.currentMode = themeMode;
    notifyListeners();
  }

  Future<void> login(String identity, String password) async {
    await _busy(() async {
      final pair = await authApi.login(identity: identity, password: password);
      accessToken = pair.accessToken;
      refreshToken = pair.refreshToken;
      await _persistTokens(pair);
      currentUser = await authApi.me();
    });
  }

  Future<void> register(String email, String login, String password) async {
    await _busy(() async {
      await authApi.register(email: email, login: login, password: password);
    });
  }

  Future<void> loadMe() async {
    await _busy(() async {
      currentUser = await authApi.me();
    });
  }

  Future<bool> refresh() async {
    final token = refreshToken;
    if (token == null || token.isEmpty) {
      return false;
    }
    try {
      final pair = await authApi.refresh(token);
      accessToken = pair.accessToken;
      refreshToken = pair.refreshToken;
      await _persistTokens(pair);
      currentUser = await authApi.me();
      notifyListeners();
      return true;
    } catch (_) {
      await _clearSession();
      notifyListeners();
      return false;
    }
  }

  Future<void> restoreSession() async {
    if (isRestoring) return;
    isRestoring = true;
    notifyListeners();
    try {
      accessToken = await storage.read(key: _accessTokenKey);
      refreshToken = await storage.read(key: _refreshTokenKey);
      if (accessToken == null || accessToken!.isEmpty) {
        accessToken = null;
        if (refreshToken != null && refreshToken!.isNotEmpty) {
          await refresh();
        }
        return;
      }
      try {
        currentUser = await authApi.me();
      } on ApiException catch (error) {
        if (error.statusCode == 401) {
          await refresh();
        } else {
          await _clearSession();
        }
      } catch (_) {
        await _clearSession();
      }
    } finally {
      isRestoring = false;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    final token = refreshToken;
    try {
      if (token != null && token.isNotEmpty) {
        await authApi.logout(token);
      }
    } finally {
      await _clearSession();
      notifyListeners();
    }
  }

  Future<void> _persistTokens(TokenPair pair) async {
    await storage.write(key: _accessTokenKey, value: pair.accessToken);
    await storage.write(key: _refreshTokenKey, value: pair.refreshToken);
  }

  Future<void> _clearSession() async {
    accessToken = null;
    refreshToken = null;
    currentUser = null;
    await storage.delete(key: _accessTokenKey);
    await storage.delete(key: _refreshTokenKey);
  }

  Future<void> _busy(Future<void> Function() action) async {
    isBusy = true;
    notifyListeners();
    try {
      await action();
    } finally {
      isBusy = false;
      notifyListeners();
    }
  }
}

class SessionScope extends InheritedNotifier<SessionController> {
  const SessionScope({
    super.key,
    required SessionController controller,
    required super.child,
  }) : super(notifier: controller);

  static SessionController of(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<SessionScope>();
    assert(scope != null, 'SessionScope not found');
    return scope!.notifier!;
  }
}

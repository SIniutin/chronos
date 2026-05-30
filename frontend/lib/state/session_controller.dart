import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../api/auth_api.dart';
import '../theme/app_theme.dart';

class SessionController extends ChangeNotifier {
  final ApiClient client;
  late final AuthApi authApi;

  String? accessToken;
  String? refreshToken;
  AppUser? currentUser;
  bool isBusy = false;
  ThemeMode themeMode = ThemeMode.dark;

  SessionController({ApiClient? apiClient}) : client = apiClient ?? ApiClient() {
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
      currentUser = await authApi.me();
      notifyListeners();
      return true;
    } catch (_) {
      accessToken = null;
      refreshToken = null;
      currentUser = null;
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    final token = refreshToken;
    try {
      if (token != null && token.isNotEmpty) {
        await authApi.logout(token);
      }
    } finally {
      accessToken = null;
      refreshToken = null;
      currentUser = null;
      notifyListeners();
    }
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

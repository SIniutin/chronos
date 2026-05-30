import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'theme/app_theme.dart';
import 'pages/home_page.dart';
import 'pages/learn_page.dart';
import 'pages/quiz_page.dart';
import 'pages/profile_page.dart';
import 'pages/splash_screen.dart';
import 'pages/onboarding_screen.dart';
import 'pages/search_page.dart';
import 'pages/auth_page.dart';
import 'state/session_controller.dart';
import 'widgets/bottom_nav.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const HistoryApp());
}

class HistoryApp extends StatefulWidget {
  const HistoryApp({super.key});

  @override
  State<HistoryApp> createState() => _HistoryAppState();
}

class _HistoryAppState extends State<HistoryApp> {
  final _session = SessionController();

  @override
  void initState() {
    super.initState();
    _session.addListener(_onSessionChanged);
    _session.restoreSession();
  }

  void _onSessionChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  @override
  void dispose() {
    _session.removeListener(_onSessionChanged);
    _session.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    AppTheme.currentMode = _session.themeMode;
    SystemChrome.setSystemUIOverlayStyle(
      SystemUiOverlayStyle(
        statusBarColor: Colors.transparent,
        statusBarIconBrightness: AppTheme.isLight ? Brightness.dark : Brightness.light,
        statusBarBrightness: AppTheme.isLight ? Brightness.light : Brightness.dark,
        systemNavigationBarColor: AppTheme.secondary,
        systemNavigationBarIconBrightness: AppTheme.isLight ? Brightness.dark : Brightness.light,
      ),
    );
    return SessionScope(
      controller: _session,
      child: MaterialApp(
        title: 'История',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        darkTheme: AppTheme.darkTheme,
        themeMode: _session.themeMode,
        home: const AppFlow(),
      ),
    );
  }
}

class AppFlow extends StatefulWidget {
  const AppFlow({super.key});

  @override
  State<AppFlow> createState() => _AppFlowState();
}

class _AppFlowState extends State<AppFlow> {
  int _step = 0; // 0=splash, 1=onboarding, 2=auth/main

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    if (session.isRestoring) {
      return const Scaffold(
        backgroundColor: AppTheme.primary,
        body: Center(child: CircularProgressIndicator(color: AppTheme.accent)),
      );
    }
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 400),
      child: switch (_step) {
        0 => SplashScreen(key: const ValueKey(0), onDone: () => setState(() => _step = 1)),
        1 => OnboardingScreen(key: const ValueKey(1), onDone: () => setState(() => _step = 2)),
        _ => session.isAuthenticated
            ? const MainScreen(key: ValueKey(2))
            : AuthPage(key: const ValueKey(3), onAuthenticated: () => setState(() {})),
      },
    );
  }
}

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _currentIndex = 0;

  void _setPage(int index) {
    setState(() => _currentIndex = index);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: _currentIndex == 0
          ? AppBar(
              backgroundColor: AppTheme.primary,
              elevation: 0,
              title: Row(
                children: [
                  const Text('📜', style: TextStyle(fontSize: 22)),
                  const SizedBox(width: 8),
                  Text(
                    'История',
                    style: TextStyle(
                      fontFamily: 'Playfair Display',
                      color: AppTheme.textPrimary,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              actions: [
                IconButton(
                  icon: Icon(Icons.search, color: AppTheme.textSecondary),
                  onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const SearchPage())),
                ),
                IconButton(
                  icon: Icon(Icons.notifications_outlined, color: AppTheme.textSecondary),
                  onPressed: () {},
                ),
              ],
            )
          : _currentIndex == 1
              ? AppBar(
                  backgroundColor: AppTheme.primary,
                  elevation: 0,
                  title: const Text('Учиться'),
                )
              : _currentIndex == 2
                  ? null
                  : AppBar(
                      backgroundColor: AppTheme.primary,
                      elevation: 0,
                      title: const Text('Профиль'),
                    ),
      body: IndexedStack(
        index: _currentIndex,
        children: [
          HomePage(
            onLearnTap: () => _setPage(1),
            onQuizTap: () => _setPage(2),
          ),
          const LearnPage(),
          const QuizPage(),
          const ProfilePage(),
        ],
      ),
      bottomNavigationBar: BottomNav(
        currentIndex: _currentIndex,
        onTap: _setPage,
      ),
    );
  }
}

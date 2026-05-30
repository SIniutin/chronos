import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../api/api_client.dart';
import '../state/session_controller.dart';
import '../theme/app_theme.dart';

class AuthPage extends StatefulWidget {
  final VoidCallback onAuthenticated;

  const AuthPage({super.key, required this.onAuthenticated});

  @override
  State<AuthPage> createState() => _AuthPageState();
}

class _AuthPageState extends State<AuthPage> {
  bool _isLogin = true;
  String? _error;
  String? _notice;

  final _identity = TextEditingController();
  final _email = TextEditingController();
  final _login = TextEditingController();
  final _password = TextEditingController();

  @override
  void dispose() {
    _identity.dispose();
    _email.dispose();
    _login.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _error = null;
      _notice = null;
    });
    final session = SessionScope.of(context);
    try {
      if (_isLogin) {
        await session.login(_identity.text.trim(), _password.text);
        widget.onAuthenticated();
      } else {
        await session.register(_email.text.trim(), _login.text.trim(), _password.text);
        setState(() {
          _isLogin = true;
          _identity.text = _login.text.trim();
          _notice = 'Аккаунт создан. Теперь войди с логином и паролем.';
        });
      }
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } catch (_) {
      setState(() => _error = 'Не удалось связаться с сервером');
    }
  }

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    return Scaffold(
      backgroundColor: AppTheme.primary,
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 36),
              Text(
                'История',
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 38,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                _isLogin ? 'Войди, чтобы продолжить обучение' : 'Создай аккаунт ученика',
                style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 16),
              ),
              const SizedBox(height: 32),
              Row(
                children: [
                  _ModeButton(label: 'Вход', active: _isLogin, onTap: () => setState(() => _isLogin = true)),
                  const SizedBox(width: 10),
                  _ModeButton(label: 'Регистрация', active: !_isLogin, onTap: () => setState(() => _isLogin = false)),
                ],
              ),
              const SizedBox(height: 24),
              if (_isLogin)
                _Input(controller: _identity, label: 'Email или login', icon: Icons.person_outline)
              else ...[
                _Input(controller: _email, label: 'Email', icon: Icons.mail_outline, keyboardType: TextInputType.emailAddress),
                const SizedBox(height: 14),
                _Input(controller: _login, label: 'Login', icon: Icons.alternate_email),
              ],
              const SizedBox(height: 14),
              _Input(controller: _password, label: 'Пароль', icon: Icons.lock_outline, obscure: true),
              if (_error != null) ...[
                const SizedBox(height: 14),
                Text(_error!, style: GoogleFonts.lato(color: AppTheme.error, fontWeight: FontWeight.bold)),
              ],
              if (_notice != null) ...[
                const SizedBox(height: 14),
                Text(_notice!, style: GoogleFonts.lato(color: AppTheme.success, fontWeight: FontWeight.bold)),
              ],
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: session.isBusy ? null : _submit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppTheme.accent,
                    foregroundColor: AppTheme.onAccent,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  child: Text(
                    session.isBusy ? 'Подождите...' : (_isLogin ? 'Войти' : 'Зарегистрироваться'),
                    style: GoogleFonts.lato(fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ModeButton extends StatelessWidget {
  final String label;
  final bool active;
  final VoidCallback onTap;

  const _ModeButton({required this.label, required this.active, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: OutlinedButton(
        onPressed: onTap,
        style: OutlinedButton.styleFrom(
          backgroundColor: active ? AppTheme.accent.withOpacity(0.16) : AppTheme.surface,
          side: BorderSide(color: active ? AppTheme.accent : AppTheme.cardBg),
          padding: const EdgeInsets.symmetric(vertical: 14),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
        ),
        child: Text(label, style: GoogleFonts.lato(color: active ? AppTheme.accent : AppTheme.textSecondary)),
      ),
    );
  }
}

class _Input extends StatelessWidget {
  final TextEditingController controller;
  final String label;
  final IconData icon;
  final bool obscure;
  final TextInputType? keyboardType;

  const _Input({
    required this.controller,
    required this.label,
    required this.icon,
    this.obscure = false,
    this.keyboardType,
  });

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      obscureText: obscure,
      keyboardType: keyboardType,
      style: GoogleFonts.lato(color: AppTheme.textPrimary),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: GoogleFonts.lato(color: AppTheme.textSecondary),
        prefixIcon: Icon(icon, color: AppTheme.textSecondary),
        filled: true,
        fillColor: AppTheme.surface,
        enabledBorder: OutlineInputBorder(
          borderSide: BorderSide(color: AppTheme.cardBg),
          borderRadius: BorderRadius.circular(14),
        ),
        focusedBorder: OutlineInputBorder(
          borderSide: const BorderSide(color: AppTheme.accent),
          borderRadius: BorderRadius.circular(14),
        ),
      ),
    );
  }
}

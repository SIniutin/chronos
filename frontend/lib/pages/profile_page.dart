import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../api/content_api.dart';
import '../api/gamification_api.dart';
import '../api/progress_api.dart';
import '../theme/app_theme.dart';
import '../repositories/content_repository.dart';
import '../state/session_controller.dart';
import '../widgets/responsive_text.dart';
import 'admin_panel_page.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    AppTheme.currentMode = session.themeMode;
    final user = session.currentUser;
    return FutureBuilder<_ProfileData>(
      future: _loadProfile(session),
      builder: (context, snapshot) {
        final profile = snapshot.data?.profile;
        final completedLessons = snapshot.data?.completedLessons ?? 0;
        final level = 'Уровень ${profile?.level ?? 1}';
        final totalXp = profile?.totalXp ?? 0;
        final streak = profile?.currentStreak ?? 0;
        final levelProgress = totalXp % 100;
        final achievements = profile?.achievements ?? const <AchievementDto>[];
        return SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 10),

              // Profile header
              Center(
                child: Column(
                  children: [
                    Stack(
                      children: [
                        Container(
                          width: 90,
                          height: 90,
                          decoration: BoxDecoration(
                            gradient: const LinearGradient(
                              colors: [AppTheme.accent, Color(0xFFFF6B35)],
                            ),
                            shape: BoxShape.circle,
                            boxShadow: [
                              BoxShadow(
                                color: AppTheme.accent.withValues(alpha: 0.4),
                                blurRadius: 16,
                                offset: const Offset(0, 6),
                              ),
                            ],
                          ),
                          child: const Center(
                            child: Text('👤', style: TextStyle(fontSize: 40)),
                          ),
                        ),
                        Positioned(
                          bottom: 0,
                          right: 0,
                          child: Container(
                            width: 28,
                            height: 28,
                            decoration: const BoxDecoration(
                              color: AppTheme.accent,
                              shape: BoxShape.circle,
                            ),
                            child: Icon(Icons.edit,
                                color: AppTheme.onAccent, size: 14),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 14),
                    Text(
                      user?.login ?? 'Историк',
                      style: GoogleFonts.playfairDisplay(
                        color: AppTheme.textPrimary,
                        fontSize: 22,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 14, vertical: 5),
                      decoration: BoxDecoration(
                        color: AppTheme.accent.withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(
                            color: AppTheme.accent.withValues(alpha: 0.4)),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Text('🏅', style: TextStyle(fontSize: 14)),
                          const SizedBox(width: 6),
                          Text(
                            level,
                            style: GoogleFonts.lato(
                              color: AppTheme.accent,
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 28),

              // Level progress
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: AppTheme.surface,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: AppTheme.cardBg),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Wrap(
                      spacing: 12,
                      runSpacing: 4,
                      alignment: WrapAlignment.spaceBetween,
                      children: [
                        ResponsiveText(
                          level,
                          style: GoogleFonts.playfairDisplay(
                            color: AppTheme.textPrimary,
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        ResponsiveText(
                          '$levelProgress%',
                          style: GoogleFonts.lato(
                            color: AppTheme.accent,
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(6),
                      child: LinearProgressIndicator(
                        value: levelProgress / 100,
                        backgroundColor: AppTheme.cardBg,
                        valueColor: const AlwaysStoppedAnimation<Color>(
                            AppTheme.accent),
                        minHeight: 10,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'До следующего уровня: ${100 - levelProgress} XP',
                      style: GoogleFonts.lato(
                          color: AppTheme.textSecondary, fontSize: 12),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 20),

              // Stats grid
              Text(
                'Статистика',
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 14),
              GridView.count(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                crossAxisCount: 2,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 1.5,
                children: [
                  _StatTile(
                      emoji: '🔥',
                      value: '$streak',
                      label: 'Дней подряд',
                      color: const Color(0xFFFF6B35)),
                  _StatTile(
                      emoji: '⭐',
                      value: '$totalXp',
                      label: 'XP всего',
                      color: AppTheme.accent),
                  _StatTile(
                      emoji: '📚',
                      value: '$completedLessons',
                      label: 'Уроков завершено',
                      color: const Color(0xFF5C7AEA)),
                  _StatTile(
                      emoji: '🏆',
                      value: '$completedLessons',
                      label: 'Квизов пройдено',
                      color: const Color(0xFF2ECC71)),
                ],
              ),

              const SizedBox(height: 24),

              // Achievements
              Text(
                'Достижения',
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 14),
              if (achievements.isEmpty)
                const _EmptyAchievements()
              else
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: achievements
                      .map((achievement) => _Achievement(
                            emoji: _achievementEmoji(achievement.code),
                            title: achievement.title,
                            earned: true,
                          ))
                      .toList(),
                ),

              const SizedBox(height: 24),

              // Settings
              Text(
                'Настройки',
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 14),
              Container(
                decoration: BoxDecoration(
                  color: AppTheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: AppTheme.cardBg),
                ),
                child: Column(
                  children: [
                    _SettingRow(
                        icon: Icons.notifications_outlined,
                        label: 'Уведомления',
                        onTap: () {}),
                    _Divider(),
                    _SettingRow(
                        icon: Icons.language_outlined,
                        label: 'Язык',
                        trailing: 'Русский',
                        onTap: () {}),
                    _Divider(),
                    _ThemeSwitchRow(
                      value: session.isLightTheme,
                      onChanged: session.setLightTheme,
                    ),
                    _Divider(),
                    if (user?.canOpenPanel == true) ...[
                      _SettingRow(
                        icon: Icons.admin_panel_settings_outlined,
                        label: 'Панель управления',
                        trailing: user?.role,
                        onTap: () => Navigator.push(
                          context,
                          MaterialPageRoute(
                              builder: (_) => const AdminPanelPage()),
                        ),
                      ),
                      _Divider(),
                    ],
                    _SettingRow(
                        icon: Icons.help_outline,
                        label: 'Помощь',
                        onTap: () {}),
                    _Divider(),
                    _SettingRow(
                        icon: Icons.logout,
                        label: 'Выйти',
                        color: AppTheme.wrong,
                        onTap: session.logout),
                  ],
                ),
              ),

              const SizedBox(height: 30),
            ],
          ),
        );
      },
    );
  }

  String _achievementEmoji(String code) => switch (code) {
        'perfect_session' => '🧠',
        'streak_3' => '🔥',
        _ => '🏛️',
      };

  Future<_ProfileData> _loadProfile(SessionController session) async {
    GamificationProfileDto? profile;
    if (session.isAuthenticated) {
      try {
        profile = await GamificationApi(session.client).getProfile();
      } catch (_) {
        profile = null;
      }
    }
    var completedLessons = 0;
    try {
      final catalog = await ContentRepository(
        ContentApi(session.client),
        progressApi:
            session.isAuthenticated ? ProgressApi(session.client) : null,
        allowFallback: false,
      ).loadCatalog();
      completedLessons =
          catalog.lessons.where((lesson) => lesson.isCompleted).length;
    } catch (_) {}
    return _ProfileData(profile: profile, completedLessons: completedLessons);
  }
}

class _ProfileData {
  final GamificationProfileDto? profile;
  final int completedLessons;

  const _ProfileData({required this.profile, required this.completedLessons});
}

class _EmptyAchievements extends StatelessWidget {
  const _EmptyAchievements();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Text(
        'Пока нет достижений',
        style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 14),
        textAlign: TextAlign.center,
      ),
    );
  }
}

class _StatTile extends StatelessWidget {
  final String emoji;
  final String value;
  final String label;
  final Color color;

  const _StatTile({
    required this.emoji,
    required this.value,
    required this.label,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(emoji, style: const TextStyle(fontSize: 22)),
          const SizedBox(height: 4),
          Text(
            value,
            style: GoogleFonts.playfairDisplay(
              color: AppTheme.textPrimary,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          Text(
            label,
            style:
                GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 11),
          ),
        ],
      ),
    );
  }
}

class _Achievement extends StatelessWidget {
  final String emoji;
  final String title;
  final bool earned;

  const _Achievement(
      {required this.emoji, required this.title, required this.earned});

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: earned ? 1.0 : 0.4,
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 180),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: earned
                ? AppTheme.accent.withValues(alpha: 0.12)
                : AppTheme.surface,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: earned
                  ? AppTheme.accent.withValues(alpha: 0.4)
                  : AppTheme.cardBg,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(emoji, style: const TextStyle(fontSize: 18)),
              const SizedBox(width: 6),
              Flexible(
                child: Text(
                  title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: GoogleFonts.lato(
                    color:
                        earned ? AppTheme.textPrimary : AppTheme.textSecondary,
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
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

class _SettingRow extends StatelessWidget {
  final IconData icon;
  final String label;
  final String? trailing;
  final Color? color;
  final VoidCallback onTap;

  const _SettingRow({
    required this.icon,
    required this.label,
    this.trailing,
    this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final c = color ?? AppTheme.textPrimary;
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        child: Row(
          children: [
            Icon(icon, color: c, size: 20),
            const SizedBox(width: 16),
            Expanded(
              child: ResponsiveText(
                label,
                style: GoogleFonts.lato(color: c, fontSize: 15),
              ),
            ),
            if (trailing != null)
              ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 130),
                child: ResponsiveText(
                  trailing!,
                  textAlign: TextAlign.end,
                  style: GoogleFonts.lato(
                      color: AppTheme.textSecondary, fontSize: 13),
                ),
              )
            else
              Icon(Icons.arrow_forward_ios,
                  color: AppTheme.textSecondary, size: 14),
          ],
        ),
      ),
    );
  }
}

class _ThemeSwitchRow extends StatelessWidget {
  final bool value;
  final ValueChanged<bool> onChanged;

  const _ThemeSwitchRow({required this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          const Icon(Icons.light_mode_outlined,
              color: AppTheme.accent, size: 20),
          const SizedBox(width: 16),
          Expanded(
            child: Text('Светлая тема',
                style: GoogleFonts.lato(
                    color: AppTheme.textPrimary, fontSize: 15)),
          ),
          Switch(
            value: value,
            onChanged: onChanged,
            activeThumbColor: AppTheme.accent,
          ),
        ],
      ),
    );
  }
}

class _Divider extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Divider(
        height: 1, color: AppTheme.cardBg, indent: 16, endIndent: 16);
  }
}

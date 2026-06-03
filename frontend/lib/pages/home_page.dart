import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../api/content_api.dart';
import '../api/gamification_api.dart';
import '../api/progress_api.dart';
import '../api/recommendation_api.dart';
import '../theme/app_theme.dart';
import '../models/models.dart';
import '../repositories/content_repository.dart';
import '../state/session_controller.dart';
import '../widgets/responsive_text.dart';
import 'timeline_page.dart';

class HomePage extends StatelessWidget {
  final VoidCallback onLearnTap;
  final VoidCallback onQuizTap;

  const HomePage({
    super.key,
    required this.onLearnTap,
    required this.onQuizTap,
  });

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    AppTheme.currentMode = session.themeMode;
    return FutureBuilder<_HomeData>(
      future: _loadHome(context),
      builder: (context, snapshot) {
        final data = snapshot.data ?? _HomeData.fallback();
        final stats = data.stats;
        final featured = data.featuredLesson;
        final fact = data.fact;
        return SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 10),
              // Header
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        ResponsiveText(
                          'Привет, Историк! 👋',
                          style: GoogleFonts.lato(
                            color: AppTheme.textSecondary,
                            fontSize: 14,
                          ),
                        ),
                        const SizedBox(height: 4),
                        ResponsiveText(
                          'Продолжай учиться',
                          style: GoogleFonts.playfairDisplay(
                            color: AppTheme.textPrimary,
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  // Streak badge
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 14, vertical: 10),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFFFF6B35), Color(0xFFFF9A3C)],
                      ),
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFFFF6B35).withValues(alpha: 0.4),
                          blurRadius: 12,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: Row(
                      children: [
                        const Text('🔥', style: TextStyle(fontSize: 18)),
                        const SizedBox(width: 4),
                        Text(
                          '${stats.streak}',
                          style: GoogleFonts.lato(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 24),

              // Stats row
              Row(
                children: [
                  _StatCard(
                    value: '${stats.totalPoints}',
                    label: 'Очков',
                    icon: '⭐',
                    color: AppTheme.accent,
                  ),
                  const SizedBox(width: 12),
                  _StatCard(
                    value: '${stats.lessonsCompleted}',
                    label: 'Уроков',
                    icon: '📚',
                    color: const Color(0xFF5C7AEA),
                  ),
                  const SizedBox(width: 12),
                  _StatCard(
                    value: '${stats.quizzesPassed}',
                    label: 'Квизов',
                    icon: '🏆',
                    color: const Color(0xFF2ECC71),
                  ),
                ],
              ),

              const SizedBox(height: 28),

              // Daily goal banner
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: AppTheme.dailyGoalGradient,
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: AppTheme.accent.withValues(alpha: 0.3),
                    width: 1,
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: AppTheme.accent.withValues(alpha: 0.2),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Text('🎯', style: TextStyle(fontSize: 28)),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Дневная цель',
                            style: GoogleFonts.playfairDisplay(
                              color: AppTheme.textPrimary,
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 6),
                          ClipRRect(
                            borderRadius: BorderRadius.circular(4),
                            child: LinearProgressIndicator(
                              value: stats.levelProgress / 100,
                              backgroundColor:
                                  AppTheme.primary.withValues(alpha: 0.5),
                              valueColor: const AlwaysStoppedAnimation<Color>(
                                  AppTheme.accent),
                              minHeight: 6,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            '${stats.levelProgress}% выполнено',
                            style: GoogleFonts.lato(
                              color: AppTheme.textSecondary,
                              fontSize: 12,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 28),

              // Continue learning
              Text(
                'Продолжи обучение',
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 16),

              // Featured lesson
              GestureDetector(
                onTap: onLearnTap,
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: AppTheme.featuredLessonGradient,
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    borderRadius: BorderRadius.circular(20),
                    boxShadow: [
                      BoxShadow(
                        color: AppTheme.isLight
                            ? const Color(0xFFD6A35D).withValues(alpha: 0.28)
                            : Colors.brown.withValues(alpha: 0.4),
                        blurRadius: 16,
                        offset: const Offset(0, 6),
                      ),
                    ],
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Flexible(
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 10, vertical: 4),
                              decoration: BoxDecoration(
                                color: AppTheme.isLight
                                    ? AppTheme.surface.withValues(alpha: 0.75)
                                    : Colors.white.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: ResponsiveText(
                                featured == null
                                    ? '📚 Обучение'
                                    : '📚 ${featured.title}',
                                style: GoogleFonts.lato(
                                  color: AppTheme.isLight
                                      ? AppTheme.textPrimary
                                      : Colors.white,
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Icon(Icons.play_circle_filled,
                              color: AppTheme.isLight
                                  ? AppTheme.textPrimary
                                  : Colors.white,
                              size: 32),
                        ],
                      ),
                      const SizedBox(height: 12),
                      ResponsiveText(
                        featured?.title ?? 'Продолжи обучение',
                        style: GoogleFonts.playfairDisplay(
                          color: AppTheme.isLight
                              ? AppTheme.textPrimary
                              : Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 6),
                      ResponsiveText(
                        featured?.description ??
                            'Открой следующий доступный урок',
                        style: GoogleFonts.lato(
                          color: AppTheme.isLight
                              ? AppTheme.textSecondary
                              : Colors.white.withValues(alpha: 0.8),
                          fontSize: 13,
                        ),
                      ),
                      const SizedBox(height: 14),
                      Wrap(
                        spacing: 12,
                        runSpacing: 6,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        children: [
                          Icon(Icons.timer_outlined,
                              color: AppTheme.isLight
                                  ? AppTheme.textSecondary
                                  : Colors.white70,
                              size: 14),
                          ResponsiveText(featured?.duration ?? '—',
                              style: GoogleFonts.lato(
                                  color: AppTheme.isLight
                                      ? AppTheme.textSecondary
                                      : Colors.white70,
                                  fontSize: 12)),
                          Icon(Icons.signal_cellular_alt,
                              color: AppTheme.isLight
                                  ? AppTheme.textSecondary
                                  : Colors.white70,
                              size: 14),
                          ResponsiveText(featured?.difficulty ?? 'Следующий',
                              style: GoogleFonts.lato(
                                  color: AppTheme.isLight
                                      ? AppTheme.textSecondary
                                      : Colors.white70,
                                  fontSize: 12)),
                        ],
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 28),

              // Quick actions
              Row(
                children: [
                  Expanded(
                    child: _QuickAction(
                      icon: '⚡',
                      title: 'Быстрый квиз',
                      subtitle: '5 вопросов',
                      color: const Color(0xFF5C7AEA),
                      onTap: onQuizTap,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _QuickAction(
                      icon: '⏳',
                      title: 'Лента времени',
                      subtitle: 'Ключевые события',
                      color: const Color(0xFF2ECC71),
                      onTap: () => Navigator.push(
                          context,
                          MaterialPageRoute(
                              builder: (_) => const TimelinePage())),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 28),

              // Today's fact
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: AppTheme.surface,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: AppTheme.cardBg,
                    width: 1,
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Text('💡', style: TextStyle(fontSize: 20)),
                        const SizedBox(width: 8),
                        Text(
                          'Факт дня',
                          style: GoogleFonts.playfairDisplay(
                            color: AppTheme.accent,
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Text(
                      fact,
                      style: GoogleFonts.lato(
                        color: AppTheme.textPrimary,
                        fontSize: 14,
                        height: 1.6,
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 20),
            ],
          ),
        );
      },
    );
  }

  Future<_HomeData> _loadHome(BuildContext context) async {
    final session = SessionScope.of(context);
    final content = ContentRepository(
      ContentApi(session.client),
      progressApi: session.isAuthenticated ? ProgressApi(session.client) : null,
      allowFallback: false,
    );
    final catalog = await content.loadCatalog();
    GamificationProfileDto? profile;
    if (session.isAuthenticated) {
      try {
        profile = await GamificationApi(session.client).getProfile();
      } catch (_) {
        profile = null;
      }
    }
    Lesson? featured;
    if (session.isAuthenticated && catalog.courseIds.isNotEmpty) {
      try {
        final rec = await RecommendationApi(session.client)
            .getNext(catalog.courseIds.first);
        for (final lesson in catalog.lessons) {
          if (lesson.backendSkillId == rec?.skillId) {
            featured = lesson;
            break;
          }
        }
      } catch (_) {
        featured = null;
      }
    }
    featured ??= catalog.lessons.isNotEmpty ? catalog.lessons.first : null;
    final completed =
        catalog.lessons.where((lesson) => lesson.isCompleted).length;
    final stats = _HomeStats(
      streak: profile?.currentStreak ?? 0,
      totalPoints: profile?.totalXp ?? 0,
      lessonsCompleted: completed,
      quizzesPassed: completed,
      levelProgress: profile == null ? 0 : profile.totalXp % 100,
    );
    return _HomeData(
        stats: stats,
        featuredLesson: featured,
        fact: _dailyFact(catalog.lessons));
  }

  String _dailyFact(List<Lesson> lessons) {
    final facts = lessons
        .expand((lesson) => lesson.facts)
        .where((fact) => fact.trim().isNotEmpty)
        .toList();
    if (facts.isEmpty) return 'Факт появится после наполнения уроков.';
    final now = DateTime.now();
    final dayKey = DateTime(now.year, now.month, now.day)
        .difference(DateTime(2020))
        .inDays;
    return facts[dayKey % facts.length];
  }
}

class _HomeData {
  final _HomeStats stats;
  final Lesson? featuredLesson;
  final String fact;

  const _HomeData(
      {required this.stats, required this.featuredLesson, required this.fact});

  factory _HomeData.fallback() => const _HomeData(
        stats: _HomeStats(
          streak: 0,
          totalPoints: 0,
          lessonsCompleted: 0,
          quizzesPassed: 0,
          levelProgress: 0,
        ),
        featuredLesson: null,
        fact: 'Факт появится после загрузки опубликованных уроков.',
      );
}

class _HomeStats {
  final int streak;
  final int totalPoints;
  final int lessonsCompleted;
  final int quizzesPassed;
  final int levelProgress;

  const _HomeStats({
    required this.streak,
    required this.totalPoints,
    required this.lessonsCompleted,
    required this.quizzesPassed,
    required this.levelProgress,
  });
}

class _StatCard extends StatelessWidget {
  final String value;
  final String label;
  final String icon;
  final Color color;

  const _StatCard({
    required this.value,
    required this.label,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withValues(alpha: 0.3), width: 1),
        ),
        child: Column(
          children: [
            Text(icon, style: const TextStyle(fontSize: 20)),
            const SizedBox(height: 6),
            ResponsiveText(
              value,
              textAlign: TextAlign.center,
              style: GoogleFonts.playfairDisplay(
                color: AppTheme.textPrimary,
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            ResponsiveText(
              label,
              textAlign: TextAlign.center,
              style: GoogleFonts.lato(
                color: AppTheme.textSecondary,
                fontSize: 11,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _QuickAction extends StatelessWidget {
  final String icon;
  final String title;
  final String subtitle;
  final Color color;
  final VoidCallback onTap;

  const _QuickAction({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withValues(alpha: 0.4), width: 1),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(icon, style: const TextStyle(fontSize: 28)),
            const SizedBox(height: 8),
            ResponsiveText(
              title,
              style: GoogleFonts.playfairDisplay(
                color: AppTheme.textPrimary,
                fontSize: 14,
                fontWeight: FontWeight.bold,
              ),
            ),
            ResponsiveText(
              subtitle,
              style: GoogleFonts.lato(
                color: AppTheme.textSecondary,
                fontSize: 12,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

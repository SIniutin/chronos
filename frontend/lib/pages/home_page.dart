import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../data/app_data.dart';
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
    const stats = AppData.userStats;
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
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Привет, Историк! 👋',
                    style: GoogleFonts.lato(
                      color: AppTheme.textSecondary,
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Продолжай учиться',
                    style: GoogleFonts.playfairDisplay(
                      color: AppTheme.textPrimary,
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              // Streak badge
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: const [Color(0xFFFF6B35), Color(0xFFFF9A3C)],
                  ),
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFFFF6B35).withOpacity(0.4),
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
                color: AppTheme.accent.withOpacity(0.3),
                width: 1,
              ),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppTheme.accent.withOpacity(0.2),
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
                          backgroundColor: AppTheme.primary.withOpacity(0.5),
                          valueColor: const AlwaysStoppedAnimation<Color>(AppTheme.accent),
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
                    color: AppTheme.isLight ? const Color(0xFFD6A35D).withOpacity(0.28) : Colors.brown.withOpacity(0.4),
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
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                        decoration: BoxDecoration(
                          color: AppTheme.isLight ? AppTheme.surface.withOpacity(0.75) : Colors.white.withOpacity(0.2),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Text(
                          '👑 Николай II',
                          style: GoogleFonts.lato(
                            color: AppTheme.isLight ? AppTheme.textPrimary : Colors.white,
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                      Icon(Icons.play_circle_filled, color: AppTheme.isLight ? AppTheme.textPrimary : Colors.white, size: 32),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Text(
                    'Воцарение Николая II',
                    style: GoogleFonts.playfairDisplay(
                      color: AppTheme.isLight ? AppTheme.textPrimary : Colors.white,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    'Как началось правление последнего российского императора',
                    style: GoogleFonts.lato(
                      color: AppTheme.isLight ? AppTheme.textSecondary : Colors.white.withOpacity(0.8),
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(height: 14),
                  Row(
                    children: [
                      Icon(Icons.timer_outlined, color: AppTheme.isLight ? AppTheme.textSecondary : Colors.white70, size: 14),
                      const SizedBox(width: 4),
                      Text('8 мин', style: GoogleFonts.lato(color: AppTheme.isLight ? AppTheme.textSecondary : Colors.white70, fontSize: 12)),
                      const SizedBox(width: 16),
                      Icon(Icons.signal_cellular_alt, color: AppTheme.isLight ? AppTheme.textSecondary : Colors.white70, size: 14),
                      const SizedBox(width: 4),
                      Text('Лёгкий', style: GoogleFonts.lato(color: AppTheme.isLight ? AppTheme.textSecondary : Colors.white70, fontSize: 12)),
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
                  onTap: () => Navigator.push(context,
                    MaterialPageRoute(builder: (_) => const TimelinePage())),
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
                  'Манифест 17 октября 1905 года впервые закрепил гражданские свободы и создание законодательной Государственной думы.',
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
  }
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
          color: color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withOpacity(0.3), width: 1),
        ),
        child: Column(
          children: [
            Text(icon, style: const TextStyle(fontSize: 20)),
            const SizedBox(height: 6),
            Text(
              value,
              style: GoogleFonts.playfairDisplay(
                color: AppTheme.textPrimary,
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              label,
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
          color: color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withOpacity(0.4), width: 1),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(icon, style: const TextStyle(fontSize: 28)),
            const SizedBox(height: 8),
            Text(
              title,
              style: GoogleFonts.playfairDisplay(
                color: AppTheme.textPrimary,
                fontSize: 14,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
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

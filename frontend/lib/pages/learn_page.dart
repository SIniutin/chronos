import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../models/models.dart';
import '../api/content_api.dart';
import '../api/progress_api.dart';
import '../api/recommendation_api.dart';
import '../repositories/content_repository.dart';
import '../state/session_controller.dart';
import '../widgets/responsive_text.dart';
import 'lesson_page.dart';

class LearnPage extends StatefulWidget {
  const LearnPage({super.key});

  @override
  State<LearnPage> createState() => _LearnPageState();
}

class _LearnPageState extends State<LearnPage> {
  late Future<CatalogSnapshot> _catalog;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final session = SessionScope.of(context);
    _catalog = ContentRepository(
      ContentApi(session.client),
      progressApi: session.isAuthenticated ? ProgressApi(session.client) : null,
      allowFallback: false,
    ).loadCatalog();
  }

  void _reloadCatalog() {
    final session = SessionScope.of(context);
    setState(() {
      _catalog = ContentRepository(
        ContentApi(session.client),
        progressApi: session.isAuthenticated ? ProgressApi(session.client) : null,
        allowFallback: false,
      ).loadCatalog();
    });
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<CatalogSnapshot>(
      future: _catalog,
      builder: (context, snapshot) {
        final catalog = snapshot.data;
        if (catalog == null && snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator(color: AppTheme.accent));
        }
        final eras = catalog?.eras ?? const <HistoryEra>[];
        final lessons = catalog?.lessons ?? const <Lesson>[];
        final courseId = catalog?.courseIds.isNotEmpty == true ? catalog!.courseIds.first : null;
        final session = SessionScope.of(context);
        return SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 10),
              Text(
                'Исторические эпохи',
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 26,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Путешествие сквозь время',
                style: GoogleFonts.lato(
                  color: AppTheme.textSecondary,
                  fontSize: 14,
                ),
              ),
              const SizedBox(height: 24),
              Container(
                decoration: BoxDecoration(
                  color: AppTheme.surface,
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: AppTheme.cardBg, width: 1),
                ),
                child: TextField(
                  style: GoogleFonts.lato(color: AppTheme.textPrimary),
                  decoration: InputDecoration(
                    hintText: 'Поиск по эпохам...',
                    hintStyle: GoogleFonts.lato(color: AppTheme.textSecondary),
                    prefixIcon: Icon(Icons.search, color: AppTheme.textSecondary),
                    filled: false,
                    border: InputBorder.none,
                    enabledBorder: InputBorder.none,
                    focusedBorder: InputBorder.none,
                    contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              if (courseId != null && session.isAuthenticated) ...[
                _NextLessonCard(courseId: courseId, lessons: lessons, onCompleted: _reloadCatalog),
                const SizedBox(height: 20),
              ],
              ...eras.map((era) => Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: _EraCard(
                      era: era,
                      onTap: () => Navigator.push<bool>(
                        context,
                        MaterialPageRoute(
                          builder: (_) => LessonsListPage(era: era, lessons: lessons),
                        ),
                      ).then((value) {
                        if (value == true && mounted) _reloadCatalog();
                      }),
                    ),
                  )),
              if (eras.isEmpty)
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(18),
                  decoration: BoxDecoration(
                    color: AppTheme.surface,
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(color: AppTheme.cardBg),
                  ),
                  child: Text(
                    'Пока нет доступных уроков',
                    style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 14),
                    textAlign: TextAlign.center,
                  ),
                ),
              const SizedBox(height: 20),
            ],
          ),
        );
      },
    );
  }
}

class _NextLessonCard extends StatefulWidget {
  final String courseId;
  final List<Lesson> lessons;
  final VoidCallback onCompleted;

  const _NextLessonCard({required this.courseId, required this.lessons, required this.onCompleted});

  @override
  State<_NextLessonCard> createState() => _NextLessonCardState();
}

class _NextLessonCardState extends State<_NextLessonCard> {
  late Future<RecommendationDto?> _recommendation;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _recommendation = RecommendationApi(SessionScope.of(context).client).getNext(widget.courseId);
  }

  @override
  void didUpdateWidget(covariant _NextLessonCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.courseId != widget.courseId || oldWidget.lessons != widget.lessons) {
      _recommendation = RecommendationApi(SessionScope.of(context).client).getNext(widget.courseId);
    }
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<RecommendationDto?>(
      future: _recommendation,
      builder: (context, snapshot) {
        final rec = snapshot.data;
        final skillId = rec?.skillId;
        final lesson = skillId == null ? null : _findLesson(skillId);
        if (lesson == null) {
          return const SizedBox.shrink();
        }

        final title = switch (rec?.type) {
          'continue' => 'Продолжить урок',
          'review' => 'Повторить тему',
          _ => 'Следующий урок',
        };

        return GestureDetector(
          onTap: () => Navigator.push<bool>(
            context,
            MaterialPageRoute(builder: (_) => LessonPage(lesson: lesson)),
          ).then((value) {
            if (value == true) widget.onCompleted();
          }),
          child: Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
              color: AppTheme.accent.withValues(alpha: 0.12),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: AppTheme.accent.withValues(alpha: 0.35)),
            ),
            child: Row(
              children: [
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    color: AppTheme.accent.withValues(alpha: 0.2),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.auto_awesome, color: AppTheme.accent),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      ResponsiveText(
                        title,
                        style: GoogleFonts.lato(
                          color: AppTheme.accent,
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 4),
                      ResponsiveText(
                        lesson.title,
                        style: GoogleFonts.playfairDisplay(
                          color: AppTheme.textPrimary,
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
                const Icon(Icons.arrow_forward_ios, color: AppTheme.accent, size: 16),
              ],
            ),
          ),
        );
      },
    );
  }

  Lesson? _findLesson(String skillId) {
    for (final lesson in widget.lessons) {
      if (lesson.backendSkillId == skillId) {
        return lesson;
      }
    }
    return null;
  }
}

class _EraCard extends StatelessWidget {
  final HistoryEra era;
  final VoidCallback onTap;

  const _EraCard({required this.era, required this.onTap});

  Color get _accentColor {
    final map = {
      'nicholas2': AppTheme.accent,
      'witte': const Color(0xFF5C7AEA),
      'revolution1905': const Color(0xFFE74C3C),
      'workers': const Color(0xFF2ECC71),
    };
    return map[era.id] ?? AppTheme.accent;
  }

  @override
  Widget build(BuildContext context) {
    final color = _accentColor;

    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: AppTheme.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: era.lessonsCompleted > 0 ? color.withValues(alpha: 0.4) : AppTheme.cardBg,
            width: 1.5,
          ),
          boxShadow: era.lessonsCompleted > 0
              ? [BoxShadow(color: color.withValues(alpha: 0.1), blurRadius: 12, offset: const Offset(0, 4))]
              : null,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 54,
                  height: 54,
                  decoration: BoxDecoration(
                    color: color.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: Center(
                    child: Text(era.emoji, style: const TextStyle(fontSize: 28)),
                  ),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      ResponsiveText(
                        era.title,
                        style: GoogleFonts.playfairDisplay(
                          color: AppTheme.textPrimary,
                          fontSize: 17,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 2),
                      ResponsiveText(
                        era.subtitle,
                        style: GoogleFonts.lato(
                          color: AppTheme.textSecondary,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ),
                Icon(Icons.arrow_forward_ios, color: color, size: 16),
              ],
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 4,
              alignment: WrapAlignment.spaceBetween,
              children: [
                ResponsiveText(
                  era.dateRange,
                  style: GoogleFonts.lato(
                    color: AppTheme.textSecondary,
                    fontSize: 12,
                  ),
                ),
                ResponsiveText(
                  '${era.lessonsCompleted}/${era.lessonsTotal} уроков',
                  style: GoogleFonts.lato(
                    color: color,
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: era.progress,
                backgroundColor: AppTheme.cardBg,
                valueColor: AlwaysStoppedAnimation<Color>(color),
                minHeight: 6,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class LessonsListPage extends StatelessWidget {
  final HistoryEra era;
  final List<Lesson> lessons;

  const LessonsListPage({super.key, required this.era, required this.lessons});

  @override
  Widget build(BuildContext context) {
    final eraLessons = lessons.where((l) => l.eraId == era.id).toList();
    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: Text(era.title),
        backgroundColor: AppTheme.primary,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: AppTheme.accent),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text(era.emoji, style: const TextStyle(fontSize: 36)),
                const SizedBox(width: 12),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      era.subtitle,
                      style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 14),
                    ),
                    Text(
                      era.dateRange,
                      style: GoogleFonts.playfairDisplay(
                        color: AppTheme.accent,
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 24),
            ...eraLessons.map((lesson) => Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: _LessonCard(
                lesson: lesson,
                onTap: () => Navigator.push<bool>(
                  context,
                  MaterialPageRoute(builder: (_) => LessonPage(lesson: lesson)),
                ).then((value) {
                  if (value == true && context.mounted) Navigator.pop(context, true);
                }),
              ),
            )),
            if (eraLessons.isEmpty)
              Center(
                child: Padding(
                  padding: const EdgeInsets.all(40),
                  child: Column(
                    children: [
                      const Text('📜', style: TextStyle(fontSize: 48)),
                      const SizedBox(height: 12),
                      Text(
                        'Уроки скоро появятся',
                        style: GoogleFonts.playfairDisplay(
                          color: AppTheme.textSecondary,
                          fontSize: 18,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _LessonCard extends StatelessWidget {
  final Lesson lesson;
  final VoidCallback onTap;

  const _LessonCard({required this.lesson, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: lesson.isLocked ? null : onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppTheme.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: lesson.isCompleted
                ? AppTheme.correct.withValues(alpha: 0.4)
                : lesson.isLocked
                    ? AppTheme.cardBg
                    : AppTheme.accent.withValues(alpha: 0.3),
            width: 1,
          ),
        ),
        child: Row(
          children: [
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: lesson.isCompleted
                    ? AppTheme.correct.withValues(alpha: 0.2)
                    : lesson.isLocked
                        ? AppTheme.cardBg
                        : AppTheme.accent.withValues(alpha: 0.15),
                shape: BoxShape.circle,
              ),
              child: Center(
                child: Icon(
                  lesson.isCompleted
                      ? Icons.check_circle
                      : lesson.isLocked
                          ? Icons.lock_outline
                          : Icons.play_circle_outlined,
                  color: lesson.isCompleted
                      ? AppTheme.correct
                      : lesson.isLocked
                          ? AppTheme.textSecondary
                          : AppTheme.accent,
                  size: 24,
                ),
              ),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  ResponsiveText(
                    lesson.title,
                    style: GoogleFonts.playfairDisplay(
                      color: lesson.isLocked ? AppTheme.textSecondary : AppTheme.textPrimary,
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Wrap(
                    spacing: 10,
                    runSpacing: 6,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Icon(Icons.timer_outlined, color: AppTheme.textSecondary, size: 12),
                      ResponsiveText(lesson.duration, style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 12)),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(
                          color: AppTheme.cardBg,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: ResponsiveText(
                          lesson.difficulty,
                          style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 10),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

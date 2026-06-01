import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../models/models.dart';
import '../widgets/responsive_text.dart';
import 'quiz_page.dart';

class LessonPage extends StatefulWidget {
  final Lesson lesson;
  const LessonPage({super.key, required this.lesson});

  @override
  State<LessonPage> createState() => _LessonPageState();
}

class _LessonPageState extends State<LessonPage> {
  @override
  Widget build(BuildContext context) {
    final lesson = widget.lesson;
    return Scaffold(
      backgroundColor: AppTheme.primary,
      body: CustomScrollView(
        slivers: [
          // Hero app bar
          SliverAppBar(
            expandedHeight: 220,
            pinned: true,
            backgroundColor: AppTheme.primary,
            leading: IconButton(
              icon: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: AppTheme.overlayOnImage,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(Icons.arrow_back_ios, color: AppTheme.textPrimary, size: 16),
              ),
              onPressed: () => Navigator.pop(context),
            ),
            flexibleSpace: FlexibleSpaceBar(
              background: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: AppTheme.lessonHeroGradient,
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                  ),
                ),
                child: Stack(
                  children: [
                    // Decorative elements
                    Positioned(
                      right: -20,
                      top: -20,
                      child: Opacity(
                        opacity: 0.1,
                        child: Container(
                          width: 200,
                          height: 200,
                          decoration: const BoxDecoration(
                            shape: BoxShape.circle,
                            color: AppTheme.accent,
                          ),
                        ),
                      ),
                    ),
                    Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const SizedBox(height: 40),
                          const Text('🏛️', style: TextStyle(fontSize: 52)),
                          const SizedBox(height: 12),
                          Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 20),
                            child: Text(
                              lesson.title,
                              textAlign: TextAlign.center,
                              style: GoogleFonts.playfairDisplay(
                                color: AppTheme.isLight ? AppTheme.textPrimary : Colors.white,
                                fontSize: 20,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),

          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Meta info
                  Wrap(
                    spacing: 10,
                    runSpacing: 8,
                    children: [
                      _MetaChip(icon: Icons.timer_outlined, label: lesson.duration, color: AppTheme.accent),
                      _MetaChip(icon: Icons.signal_cellular_alt, label: lesson.difficulty, color: const Color(0xFF5C7AEA)),
                      if (lesson.isCompleted)
                        const _MetaChip(icon: Icons.check_circle, label: 'Пройдено', color: AppTheme.correct),
                    ],
                  ),

                  const SizedBox(height: 24),

                  Text(
                    'О чём урок',
                    style: GoogleFonts.playfairDisplay(
                      color: AppTheme.textPrimary,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    lesson.description,
                    style: GoogleFonts.lato(
                      color: AppTheme.textPrimary,
                      fontSize: 15,
                      height: 1.7,
                    ),
                  ),

                  const SizedBox(height: 28),

                  // Interesting facts
                  Text(
                    'Интересные факты',
                    style: GoogleFonts.playfairDisplay(
                      color: AppTheme.textPrimary,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),

                  ...lesson.facts.asMap().entries.map((entry) {
                    final i = entry.key;
                    final fact = entry.value;
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: _FactCard(fact: fact, index: i),
                    );
                  }),

                  const SizedBox(height: 28),

                  // Content sections
                  _ContentSection(
                    title: 'Ключевые события',
                    content: _getKeyEvents(lesson),
                  ),
                  const SizedBox(height: 20),
                  _ContentSection(
                    title: 'Историческое значение',
                    content: _getSignificance(lesson),
                  ),

                  const SizedBox(height: 36),

                  // Action button
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () async {
                        final completed = await Navigator.push<bool>(
                          context,
                          MaterialPageRoute(
                            builder: (_) => QuizPage(
                              questions: lesson.quizQuestions,
                              skillId: lesson.backendSkillId,
                            ),
                          ),
                        );
                        if (completed == true && context.mounted) {
                          Navigator.pop(context, true);
                        }
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppTheme.accent,
                        foregroundColor: AppTheme.onAccent,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                        elevation: 8,
                        shadowColor: AppTheme.accent.withValues(alpha: 0.4),
                      ),
                      child: ButtonLabel(
                        'Пройти квиз по теме →',
                        style: GoogleFonts.lato(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 30),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _getKeyEvents(Lesson lesson) {
    if (lesson.id == 'l_n2_1') {
      return '• 1894 — смерть Александра III и вступление Николая II на престол\n• 1895 — Николай II отвергает идеи участия земств в управлении\n• Конец XIX века — быстрый промышленный и культурный рост\n• Нарастание конфликта между самодержавием и обществом';
    }
    if (lesson.id == 'l_rev_1') {
      return '• 8 января 1905 — ввод войск в столицу\n• 9 января 1905 — расстрел рабочих колонн у Зимнего дворца\n• После 9 января — рост стачек и политических требований\n• События стали началом революционного кризиса 1905 года';
    }
    return '• Серия важнейших событий России начала XX века\n• Рост социальных и политических противоречий\n• Попытки реформ и сопротивление самодержавной системы\n• Последствия, которые повлияли на дальнейший ход истории России';
  }

  String _getSignificance(Lesson lesson) {
    if (lesson.id == 'l_n2_1') {
      return 'Начало правления Николая II стало точкой, где экономический подъём, общественные ожидания и приверженность самодержавию вошли в острое противоречие.';
    }
    if (lesson.id == 'l_rev_1') {
      return 'Кровавое воскресенье разрушило веру значительной части общества в царя-заступника и резко ускорило революционные процессы в стране.';
    }
    return 'Этот период показывает, как реформы, социальное напряжение и политический кризис меняли Российскую империю на рубеже XIX–XX веков.';
  }
}

class _MetaChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;

  const _MetaChip({required this.icon, required this.label, required this.color});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: color.withValues(alpha: 0.4), width: 1),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 180),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, color: color, size: 14),
            const SizedBox(width: 4),
            Flexible(
              child: ResponsiveText(
                label,
                style: GoogleFonts.lato(color: color, fontSize: 12, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _FactCard extends StatefulWidget {
  final String fact;
  final int index;
  const _FactCard({required this.fact, required this.index});

  @override
  State<_FactCard> createState() => _FactCardState();
}

class _FactCardState extends State<_FactCard> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () => setState(() => _expanded = !_expanded),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppTheme.surface,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: _expanded ? AppTheme.accent.withValues(alpha: 0.5) : AppTheme.cardBg,
          ),
        ),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                color: AppTheme.accent.withValues(alpha: 0.2),
                shape: BoxShape.circle,
              ),
              child: Center(
                child: Text(
                  '${widget.index + 1}',
                  style: GoogleFonts.lato(
                    color: AppTheme.accent,
                    fontSize: 13,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                widget.fact,
                style: GoogleFonts.lato(
                  color: AppTheme.textPrimary,
                  fontSize: 14,
                  height: 1.5,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ContentSection extends StatelessWidget {
  final String title;
  final String content;

  const _ContentSection({required this.title, required this.content});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: AppTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppTheme.cardBg, width: 1),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: GoogleFonts.playfairDisplay(
              color: AppTheme.accent,
              fontSize: 15,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            content,
            style: GoogleFonts.lato(
              color: AppTheme.textPrimary,
              fontSize: 14,
              height: 1.6,
            ),
          ),
        ],
      ),
    );
  }
}

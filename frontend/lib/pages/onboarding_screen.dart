import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../widgets/responsive_text.dart';

class OnboardingScreen extends StatefulWidget {
  final VoidCallback onDone;
  const OnboardingScreen({super.key, required this.onDone});

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final PageController _pageCtrl = PageController();
  int _current = 0;

  final List<_OnboardPage> _pages = const [
    _OnboardPage(
      emoji: '🌍',
      title: 'Изучай историю\nпо-новому',
      subtitle: 'Интерактивные уроки, увлекательные факты и квизы — всё в одном приложении',
      color: Color(0xFFE8A838),
    ),
    _OnboardPage(
      emoji: '⚡',
      title: 'Учись каждый\nдень',
      subtitle: 'Всего 10 минут в день — и ты узнаешь больше, чем за год школьных уроков',
      color: Color(0xFF5C7AEA),
    ),
    _OnboardPage(
      emoji: '🏆',
      title: 'Зарабатывай\nдостижения',
      subtitle: 'Собирай очки, поддерживай стрик и становись настоящим историком',
      color: Color(0xFF2ECC71),
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.primary,
      body: SafeArea(
        child: Column(
          children: [
            // Skip button
            Align(
              alignment: Alignment.topRight,
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: TextButton(
                  onPressed: widget.onDone,
                  child: Text(
                    'Пропустить',
                    style: GoogleFonts.lato(
                      color: AppTheme.textSecondary,
                      fontSize: 14,
                    ),
                  ),
                ),
              ),
            ),

            // Pages
            Expanded(
              child: PageView.builder(
                controller: _pageCtrl,
                onPageChanged: (i) => setState(() => _current = i),
                itemCount: _pages.length,
                itemBuilder: (_, i) => _OnboardPageWidget(page: _pages[i]),
              ),
            ),

            // Dots + button
            Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                children: [
                  // Dots
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: List.generate(
                      _pages.length,
                      (i) => AnimatedContainer(
                        duration: const Duration(milliseconds: 300),
                        margin: const EdgeInsets.symmetric(horizontal: 4),
                        width: _current == i ? 24 : 8,
                        height: 8,
                        decoration: BoxDecoration(
                          color: _current == i
                              ? _pages[_current].color
                              : AppTheme.cardBg,
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 28),

                  // Button
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () {
                        if (_current < _pages.length - 1) {
                          _pageCtrl.nextPage(
                            duration: const Duration(milliseconds: 400),
                            curve: Curves.easeInOut,
                          );
                        } else {
                          widget.onDone();
                        }
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: _pages[_current].color,
                        foregroundColor: AppTheme.onAccent,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                        elevation: 8,
                        shadowColor: _pages[_current].color.withValues(alpha: 0.4),
                      ),
                      child: ButtonLabel(
                        _current < _pages.length - 1 ? 'Далее →' : 'Начать учиться! 🚀',
                        style: GoogleFonts.lato(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
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

class _OnboardPage {
  final String emoji;
  final String title;
  final String subtitle;
  final Color color;

  const _OnboardPage({
    required this.emoji,
    required this.title,
    required this.subtitle,
    required this.color,
  });
}

class _OnboardPageWidget extends StatelessWidget {
  final _OnboardPage page;
  const _OnboardPageWidget({required this.page});

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final compact = constraints.maxWidth < 360 || constraints.maxHeight < 560;
        return SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: ConstrainedBox(
            constraints: BoxConstraints(minHeight: constraints.maxHeight),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  width: compact ? 128.0 : 160.0,
                  height: compact ? 128.0 : 160.0,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: page.color.withValues(alpha: 0.12),
                    border: Border.all(color: page.color.withValues(alpha: 0.3), width: 2),
                    boxShadow: [
                      BoxShadow(
                        color: page.color.withValues(alpha: 0.2),
                        blurRadius: compact ? 28 : 40,
                        spreadRadius: compact ? 6 : 10,
                      ),
                    ],
                  ),
                  child: Center(
                    child: Text(page.emoji, style: TextStyle(fontSize: compact ? 58.0 : 72.0)),
                  ),
                ),
                SizedBox(height: compact ? 28 : 40),
                ResponsiveText(
                  page.title,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.playfairDisplay(
                    color: AppTheme.textPrimary,
                    fontSize: compact ? 28.0 : 32.0,
                    fontWeight: FontWeight.bold,
                    height: 1.2,
                  ),
                ),
                const SizedBox(height: 16),
                ResponsiveText(
                  page.subtitle,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.lato(
                    color: AppTheme.textSecondary,
                    fontSize: compact ? 15.0 : 16.0,
                    height: 1.6,
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

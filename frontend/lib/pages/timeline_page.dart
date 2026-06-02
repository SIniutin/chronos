import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../data/timeline_events.dart';
import '../api/content_api.dart';
import '../state/session_controller.dart';
import '../theme/app_theme.dart';
import '../widgets/history_timeline.dart';
import '../widgets/responsive_text.dart';

class TimelinePage extends StatefulWidget {
  const TimelinePage({super.key});

  @override
  State<TimelinePage> createState() => _TimelinePageState();
}

class _TimelinePageState extends State<TimelinePage> {
  Future<List<TimelineEvent>>? _eventsFuture;

  static final _datePattern = RegExp(
    r'(?:(?:\d{1,2}\s+[А-Яа-яЁё]+|[А-Яа-яЁё]+(?:–|-)[А-Яа-яЁё]+|[А-Яа-яЁё]+)\s+)?(18\d{2}|19\d{2})',
  );

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final session = SessionScope.of(context);
    AppTheme.currentMode = session.themeMode;
    _eventsFuture ??= _loadEvents(ContentApi(session.client));
  }

  Future<List<TimelineEvent>> _loadEvents(ContentApi api) async {
    final collected = <_CollectedTimelineEvent>[];
    var order = 0;
    final courses = await api.listCourses();
    for (final course in courses) {
      final sections = await api.listSections(course.id);
      sections.sort((a, b) => a.position.compareTo(b.position));
      for (final section in sections) {
        final units = await api.listUnits(section.id);
        units.sort((a, b) => a.position.compareTo(b.position));
        for (final unit in units) {
          final skills = await api.listSkills(unit.id);
          skills.sort((a, b) => a.position.compareTo(b.position));
          for (final skill in skills) {
            final challenges = await api.listChallenges(skill.id);
            challenges.sort((a, b) => a.position.compareTo(b.position));
            final texts = _timelineTexts(section, unit, skill, challenges);
            for (final text in texts) {
              final event = _eventFromText(text, skill.title, order);
              if (event != null) {
                collected.add(event);
                order++;
              }
            }
          }
        }
      }
    }

    final byKey = <String, _CollectedTimelineEvent>{};
    for (final event in collected) {
      byKey.putIfAbsent(
        '${event.sortYear}|${event.event.title}|${event.event.description}',
        () => event,
      );
    }
    final events = byKey.values.toList()
      ..sort((a, b) {
        final year = a.sortYear.compareTo(b.sortYear);
        return year != 0 ? year : a.order.compareTo(b.order);
      });
    return events.map((item) => item.event).toList();
  }

  Iterable<String> _timelineTexts(
    SectionDto section,
    UnitDto unit,
    SkillDto skill,
    List<ChallengeDto> challenges,
  ) sync* {
    final sectionText = _clean(section.description);
    if (_datePattern.hasMatch(sectionText)) yield sectionText;
    for (final challenge in challenges) {
      if (challenge.status.isNotEmpty && challenge.status != 'published') {
        continue;
      }
      final payload = challenge.payload;
      if (payload is Map && payload['facts'] is List) {
        for (final fact in payload['facts'] as List) {
          final text = _clean(fact);
          if (_datePattern.hasMatch(text)) yield text;
        }
      }
      for (final raw in [
        challenge.body,
        challenge.prompt,
        challenge.explanation
      ]) {
        final text = _clean(raw);
        if (_datePattern.hasMatch(text)) yield text;
      }
    }
  }

  _CollectedTimelineEvent? _eventFromText(
    String text,
    String skillTitle,
    int order,
  ) {
    final match = _datePattern.firstMatch(text);
    if (match == null) return null;
    final year = int.tryParse(match.group(1) ?? '');
    if (year == null) return null;
    final yearLabel = _cleanYearLabel(match.group(0) ?? '$year');
    return _CollectedTimelineEvent(
      sortYear: year,
      order: order,
      event: TimelineEvent(
        year: yearLabel,
        title: _titleFromText(text, skillTitle, yearLabel),
        description: text,
        emoji: _emojiFor(text, year),
        color: _colorFor(year),
      ),
    );
  }

  String _titleFromText(String text, String fallback, String yearLabel) {
    var title = text;
    final sentenceEnd = title.indexOf('.');
    if (sentenceEnd > 18) {
      title = title.substring(0, sentenceEnd);
    }
    title = title
        .replaceFirst(RegExp('^В\\s+', caseSensitive: false), '')
        .replaceFirst(yearLabel, '')
        .replaceAll(RegExp(r'\s+'), ' ')
        .trim();
    title = title.replaceAll(RegExp(r'^[,.:;–-]+\s*'), '');
    if (title.length < 8) title = fallback;
    if (title.length > 74) title = '${title.substring(0, 71).trim()}...';
    return title;
  }

  String _cleanYearLabel(String value) {
    final cleaned = value
        .replaceFirst(RegExp('^В\\s+', caseSensitive: false), '')
        .replaceAll(RegExp(r'\s+'), ' ')
        .trim();
    if (cleaned.isEmpty) return value;
    return cleaned[0].toUpperCase() + cleaned.substring(1);
  }

  String _clean(Object? value) =>
      value?.toString().replaceAll(RegExp(r'\s+'), ' ').trim() ?? '';

  Color _colorFor(int year) {
    if (year < 1905) return const Color(0xFFE8A838);
    if (year < 1918) return const Color(0xFFE74C3C);
    if (year < 1945) return const Color(0xFF9B59B6);
    if (year < 1985) return const Color(0xFF5C7AEA);
    return const Color(0xFF2ECC71);
  }

  String _emojiFor(String text, int year) {
    final lower = text.toLowerCase();
    if (lower.contains('войн') || lower.contains('восстан')) return '⚔️';
    if (lower.contains('реформ') || lower.contains('перестрой')) return '🔄';
    if (lower.contains('дума') || lower.contains('закон')) return '🏛️';
    if (lower.contains('револю')) return '✊';
    if (year >= 1985) return '🕊️';
    return '📜';
  }

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    AppTheme.currentMode = session.themeMode;
    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: const Text('Лента истории'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: AppTheme.accent),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: FutureBuilder<List<TimelineEvent>>(
        future: _eventsFuture,
        builder: (context, snapshot) {
          final events = snapshot.data?.isNotEmpty == true
              ? snapshot.data!
              : timelineEvents;
          return SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        AppTheme.accent.withValues(alpha: 0.15),
                        AppTheme.cardBg,
                      ],
                    ),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Row(
                    children: [
                      const Text('⏳', style: TextStyle(fontSize: 28)),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Россия и СССР: ключевые события курса',
                              style: GoogleFonts.playfairDisplay(
                                color: AppTheme.textPrimary,
                                fontSize: 15,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '${events.length} событий с 1882 по 1991 год',
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
                const SizedBox(height: 12),
                const Wrap(
                  spacing: 10,
                  runSpacing: 6,
                  children: [
                    _LegendChip(
                      color: Color(0xFFE8A838),
                      label: 'Российская империя',
                    ),
                    _LegendChip(
                      color: Color(0xFFE74C3C),
                      label: 'Революция и Гражданская война',
                    ),
                    _LegendChip(
                      color: Color(0xFF9B59B6),
                      label: 'СССР до войны',
                    ),
                    _LegendChip(
                      color: Color(0xFF5C7AEA),
                      label: 'Вторая мировая война',
                    ),
                    _LegendChip(
                      color: Color(0xFF2ECC71),
                      label: 'Послевоенный СССР',
                    ),
                    _LegendChip(
                      color: Color(0xFF16A085),
                      label: '1964–1984',
                    ),
                    _LegendChip(
                      color: Color(0xFFFF8C42),
                      label: 'Перестройка и распад СССР',
                    ),
                  ],
                ),
                const SizedBox(height: 22),
                if (snapshot.connectionState == ConnectionState.waiting)
                  const Center(
                    child: Padding(
                      padding: EdgeInsets.symmetric(vertical: 36),
                      child: CircularProgressIndicator(color: AppTheme.accent),
                    ),
                  )
                else if (snapshot.hasError)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      'Показываю резервную ленту: backend сейчас недоступен.',
                      style: GoogleFonts.lato(
                        color: AppTheme.textSecondary,
                        fontSize: 13,
                      ),
                    ),
                  ),
                HistoryTimeline(events: events),
                const SizedBox(height: 20),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _CollectedTimelineEvent {
  final int sortYear;
  final int order;
  final TimelineEvent event;

  const _CollectedTimelineEvent({
    required this.sortYear,
    required this.order,
    required this.event,
  });
}

class _LegendChip extends StatelessWidget {
  final Color color;
  final String label;

  const _LegendChip({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 190),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 5),
          Flexible(
            child: ResponsiveText(
              label,
              style: GoogleFonts.lato(
                color: AppTheme.textSecondary,
                fontSize: 11,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

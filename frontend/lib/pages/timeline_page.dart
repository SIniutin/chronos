import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../data/timeline_events.dart';
import '../theme/app_theme.dart';
import '../widgets/history_timeline.dart';
import '../widgets/responsive_text.dart';

class TimelinePage extends StatelessWidget {
  const TimelinePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: const Text('Лента истории'),
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
                          '$timelineEventCount событий с 1882 по 1991 год',
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
            const SizedBox(height: 28),
            const HistoryTimeline(events: timelineEvents),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }
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

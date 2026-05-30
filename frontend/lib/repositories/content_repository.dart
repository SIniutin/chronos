import '../api/content_api.dart';
import '../data/app_data.dart';
import '../models/models.dart';

class CatalogSnapshot {
  final List<String> courseIds;
  final List<HistoryEra> eras;
  final List<Lesson> lessons;

  const CatalogSnapshot({
    this.courseIds = const [],
    required this.eras,
    required this.lessons,
  });
}

class ContentRepository {
  final ContentApi api;

  const ContentRepository(this.api);

  Future<CatalogSnapshot> loadCatalog() async {
    try {
      final courses = await api.listCourses();
      if (courses.isEmpty) {
        return _fallback();
      }

      final eras = <HistoryEra>[];
      final lessons = <Lesson>[];
      for (final course in courses) {
        final sections = await api.listSections(course.id);
        for (var i = 0; i < sections.length; i++) {
          final section = sections[i];
          final sectionLessons = <Lesson>[];
          final units = await api.listUnits(section.id);
          for (final unit in units) {
            final skills = await api.listSkills(unit.id);
            for (final skill in skills) {
              final challenges = await api.listChallenges(skill.id);
              final questions = challenges
                  .where((challenge) => challenge.type != 'theory')
                  .map(_questionFromChallenge)
                  .whereType<QuizQuestion>()
                  .toList();
              sectionLessons.add(Lesson(
                id: skill.id,
                backendSkillId: skill.id,
                backendCourseId: course.id,
                backendUnitId: unit.id,
                title: skill.title.isEmpty ? unit.title : skill.title,
                description: unit.title,
                duration: '${(challenges.length.clamp(1, 5) * 4)} мин',
                difficulty: _difficulty(challenges),
                isCompleted: false,
                isLocked: false,
                eraId: section.id,
                facts: _factsFromChallenges(challenges),
                quizQuestions: questions,
              ));
            }
          }
          eras.add(HistoryEra(
            id: section.id,
            title: section.theme,
            subtitle: section.description,
            dateRange: course.title,
            emoji: _emoji(i),
            lessonsTotal: sectionLessons.isEmpty ? 1 : sectionLessons.length,
            lessonsCompleted: 0,
            color: _color(i),
          ));
          lessons.addAll(sectionLessons);
        }
      }

      if (eras.isEmpty || lessons.isEmpty) {
        return _fallback();
      }
      return CatalogSnapshot(courseIds: courses.map((course) => course.id).toList(), eras: eras, lessons: lessons);
    } catch (_) {
      return _fallback();
    }
  }

  CatalogSnapshot _fallback() => const CatalogSnapshot(
        eras: AppData.eras,
        lessons: AppData.lessons,
      );

  QuizQuestion? _questionFromChallenge(ChallengeDto challenge) {
    final options = _options(challenge.options);
    if (challenge.prompt.isEmpty || options.isEmpty) {
      return null;
    }
    return QuizQuestion(
      question: challenge.prompt,
      options: options,
      correctIndex: _correctIndex(challenge.answers, options),
      explanation: challenge.explanation,
    );
  }

  List<String> _factsFromChallenges(List<ChallengeDto> challenges) {
    final facts = challenges
        .map((challenge) => challenge.body.isNotEmpty ? challenge.body : challenge.explanation)
        .where((text) => text.trim().isNotEmpty)
        .take(5)
        .toList();
    return facts.isEmpty ? const ['Материал появится после наполнения контента.'] : facts;
  }

  List<String> _options(dynamic raw) {
    if (raw is List) {
      return raw.map((item) {
        if (item is String) return item;
        if (item is Map && item['text'] is String) return item['text'] as String;
        return item.toString();
      }).where((text) => text.trim().isNotEmpty).toList();
    }
    return const [];
  }

  int? _correctIndex(dynamic rawAnswers, List<String> options) {
    if (rawAnswers is List && rawAnswers.isNotEmpty) {
      final answer = rawAnswers.first;
      if (answer is int && answer >= 0 && answer < options.length) {
        return answer;
      }
      if (answer is String) {
        final byText = options.indexOf(answer);
        if (byText >= 0) return byText;
        final parsed = int.tryParse(answer);
        if (parsed != null && parsed >= 0 && parsed < options.length) return parsed;
      }
    }
    return null;
  }

  String _difficulty(List<ChallengeDto> challenges) {
    if (challenges.any((challenge) => challenge.difficulty == 'hard')) return 'Сложный';
    if (challenges.any((challenge) => challenge.difficulty == 'medium')) return 'Средний';
    return 'Лёгкий';
  }

  String _emoji(int index) => const ['👑', '💰', '✊', '⚙️', '📜'][index % 5];
  String _color(int index) => const ['#E8A838', '#5C7AEA', '#E74C3C', '#2ECC71', '#F5C842'][index % 5];
}

class HistoryEra {
  final String id;
  final String title;
  final String subtitle;
  final String dateRange;
  final String emoji;
  final int lessonsTotal;
  final int lessonsCompleted;
  final String color;

  const HistoryEra({
    required this.id,
    required this.title,
    required this.subtitle,
    required this.dateRange,
    required this.emoji,
    required this.lessonsTotal,
    required this.lessonsCompleted,
    required this.color,
  });

  double get progress => lessonsCompleted / lessonsTotal;
}

class Lesson {
  final String id;
  final String? backendSkillId;
  final String? backendCourseId;
  final String? backendUnitId;
  final String title;
  final String description;
  final String duration;
  final String difficulty;
  final bool isCompleted;
  final bool isLocked;
  final String eraId;
  final List<String> facts;
  final List<QuizQuestion> quizQuestions;

  const Lesson({
    required this.id,
    this.backendSkillId,
    this.backendCourseId,
    this.backendUnitId,
    required this.title,
    required this.description,
    required this.duration,
    required this.difficulty,
    required this.isCompleted,
    required this.isLocked,
    required this.eraId,
    required this.facts,
    this.quizQuestions = const [],
  });
}

class QuizQuestion {
  final String question;
  final List<String> options;
  final int? correctIndex;
  final String explanation;

  const QuizQuestion({
    required this.question,
    required this.options,
    this.correctIndex,
    required this.explanation,
  });
}

class UserStats {
  final int streak;
  final int totalPoints;
  final int lessonsCompleted;
  final int quizzesPassed;
  final String level;
  final int levelProgress;

  const UserStats({
    required this.streak,
    required this.totalPoints,
    required this.lessonsCompleted,
    required this.quizzesPassed,
    required this.level,
    required this.levelProgress,
  });
}

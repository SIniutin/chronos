import 'content_api.dart';
import 'api_client.dart';

class LessonSessionDto {
  final String id;
  final String userId;
  final String skillId;
  final String status;
  final String startedAt;
  final String? finishedAt;

  const LessonSessionDto({
    required this.id,
    required this.userId,
    required this.skillId,
    required this.status,
    required this.startedAt,
    this.finishedAt,
  });

  factory LessonSessionDto.fromJson(Map<String, dynamic> json) => LessonSessionDto(
        id: json['id'] as String? ?? '',
        userId: json['user_id'] as String? ?? '',
        skillId: json['skill_id'] as String? ?? '',
        status: json['status'] as String? ?? '',
        startedAt: json['started_at'] as String? ?? '',
        finishedAt: json['finished_at'] as String?,
      );
}

class CurrentChallengeDto {
  final String sessionChallengeId;
  final int position;
  final ChallengeDto challenge;

  const CurrentChallengeDto({
    required this.sessionChallengeId,
    required this.position,
    required this.challenge,
  });

  factory CurrentChallengeDto.fromJson(Map<String, dynamic> json) => CurrentChallengeDto(
        sessionChallengeId: json['session_challenge_id'] as String? ?? '',
        position: json['position'] as int? ?? 0,
        challenge: ChallengeDto.fromJson(json['challenge'] as Map<String, dynamic>),
      );
}

class SubmitAnswerResultDto {
  final String attemptId;
  final bool isCorrect;
  final List<String> mistakes;
  final bool hasNext;

  const SubmitAnswerResultDto({
    required this.attemptId,
    required this.isCorrect,
    required this.mistakes,
    required this.hasNext,
  });

  factory SubmitAnswerResultDto.fromJson(Map<String, dynamic> json) => SubmitAnswerResultDto(
        attemptId: json['attempt_id'] as String? ?? '',
        isCorrect: json['is_correct'] as bool? ?? false,
        mistakes: _stringList(json['mistakes']),
        hasNext: json['has_next'] as bool? ?? false,
      );
}

class LessonSessionResultDto {
  final String sessionId;
  final int total;
  final int correct;
  final int percent;

  const LessonSessionResultDto({
    required this.sessionId,
    required this.total,
    required this.correct,
    required this.percent,
  });

  factory LessonSessionResultDto.fromJson(Map<String, dynamic> json) => LessonSessionResultDto(
        sessionId: json['session_id'] as String? ?? '',
        total: json['total'] as int? ?? 0,
        correct: json['correct'] as int? ?? 0,
        percent: json['percent'] as int? ?? 0,
      );
}

class LearningApi {
  final ApiClient _client;

  const LearningApi(this._client);

  Future<LessonSessionDto> startSession(String skillId, {int limit = 10}) async {
    final json = await _client.post('/learning/sessions', body: {
      'skill_id': skillId,
      'limit': limit,
    }, auth: true);
    return LessonSessionDto.fromJson(json as Map<String, dynamic>);
  }

  Future<CurrentChallengeDto> getCurrentChallenge(String sessionId) async {
    final json = await _client.get('/learning/sessions/$sessionId/current', auth: true);
    return CurrentChallengeDto.fromJson(json as Map<String, dynamic>);
  }

  Future<SubmitAnswerResultDto> submitAnswer(String sessionId, Object? userAnswer) async {
    final json = await _client.post('/learning/sessions/$sessionId/answer', body: {
      'user_answer': userAnswer,
    }, auth: true);
    return SubmitAnswerResultDto.fromJson(json as Map<String, dynamic>);
  }

  Future<LessonSessionResultDto> finishSession(String sessionId) async {
    final json = await _client.post('/learning/sessions/$sessionId/finish', auth: true);
    return LessonSessionResultDto.fromJson(json as Map<String, dynamic>);
  }
}

List<String> _stringList(dynamic raw) {
  if (raw is List) {
    return raw.map((item) => item.toString()).toList();
  }
  return const [];
}

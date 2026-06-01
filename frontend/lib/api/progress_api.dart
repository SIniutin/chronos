import 'api_client.dart';

class ProgressCatalogDto {
  final String courseId;
  final String courseStatus;
  final int totalLessons;
  final int availableLessons;
  final int completedLessons;
  final List<ProgressUnitDto> units;
  final List<ProgressSkillDto> skills;

  const ProgressCatalogDto({
    required this.courseId,
    required this.courseStatus,
    required this.totalLessons,
    required this.availableLessons,
    required this.completedLessons,
    required this.units,
    required this.skills,
  });

  factory ProgressCatalogDto.fromJson(Map<String, dynamic> json) => ProgressCatalogDto(
        courseId: json['course_id'] as String? ?? '',
        courseStatus: json['course_status'] as String? ?? 'available',
        totalLessons: json['total_lessons'] as int? ?? 0,
        availableLessons: json['available_lessons'] as int? ?? 0,
        completedLessons: json['completed_lessons'] as int? ?? 0,
        units: (json['units'] as List? ?? const [])
            .map((item) => ProgressUnitDto.fromJson(item as Map<String, dynamic>))
            .toList(),
        skills: (json['skills'] as List? ?? const [])
            .map((item) => ProgressSkillDto.fromJson(item as Map<String, dynamic>))
            .toList(),
      );

  Map<String, ProgressSkillDto> get skillsById => {
        for (final skill in skills) skill.skillId: skill,
      };
}

class ProgressUnitDto {
  final String unitId;
  final String status;

  const ProgressUnitDto({required this.unitId, required this.status});

  factory ProgressUnitDto.fromJson(Map<String, dynamic> json) => ProgressUnitDto(
        unitId: json['unit_id'] as String? ?? '',
        status: json['status'] as String? ?? 'locked',
      );
}

class ProgressSkillDto {
  final String skillId;
  final String unitId;
  final String status;
  final int level;
  final double mastery;
  final int correctAnswers;
  final int wrongAnswers;

  const ProgressSkillDto({
    required this.skillId,
    required this.unitId,
    required this.status,
    required this.level,
    required this.mastery,
    required this.correctAnswers,
    required this.wrongAnswers,
  });

  factory ProgressSkillDto.fromJson(Map<String, dynamic> json) => ProgressSkillDto(
        skillId: json['skill_id'] as String? ?? '',
        unitId: json['unit_id'] as String? ?? '',
        status: json['status'] as String? ?? 'locked',
        level: json['level'] as int? ?? 0,
        mastery: (json['mastery'] as num?)?.toDouble() ?? 0,
        correctAnswers: json['correct_answers'] as int? ?? 0,
        wrongAnswers: json['wrong_answers'] as int? ?? 0,
      );
}

class ProgressApi {
  final ApiClient _client;

  const ProgressApi(this._client);

  Future<ProgressCatalogDto> getCatalog(String courseId) async {
    final encoded = Uri.encodeQueryComponent(courseId);
    final json = await _client.get('/progress/catalog?course_id=$encoded', auth: true);
    return ProgressCatalogDto.fromJson(json as Map<String, dynamic>);
  }

  Future<ProgressCatalogDto> completeAllForUser(String userId) async {
    final encoded = Uri.encodeComponent(userId);
    final json = await _client.post('/admin/progress/users/$encoded/complete-all', auth: true);
    return ProgressCatalogDto.fromJson(json as Map<String, dynamic>);
  }
}

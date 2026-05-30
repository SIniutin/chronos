import 'dart:convert';

import 'api_client.dart';

class CourseDto {
  final String id;
  final String sourceLang;
  final String targetLang;
  final String title;
  final String status;

  const CourseDto({
    required this.id,
    required this.sourceLang,
    required this.targetLang,
    required this.title,
    required this.status,
  });

  factory CourseDto.fromJson(Map<String, dynamic> json) => CourseDto(
        id: json['id'] as String? ?? '',
        sourceLang: json['source_lang'] as String? ?? '',
        targetLang: json['target_lang'] as String? ?? '',
        title: json['title'] as String? ?? '',
        status: json['status'] as String? ?? '',
      );

  Map<String, dynamic> toWriteJson() => {
        'source_lang': sourceLang,
        'target_lang': targetLang,
        'title': title,
      };
}

class SectionDto {
  final String id;
  final String courseId;
  final String theme;
  final String description;
  final int position;
  final String status;

  const SectionDto({
    required this.id,
    required this.courseId,
    required this.theme,
    required this.description,
    required this.position,
    required this.status,
  });

  factory SectionDto.fromJson(Map<String, dynamic> json) => SectionDto(
        id: json['id'] as String? ?? '',
        courseId: json['course_id'] as String? ?? '',
        theme: json['theme'] as String? ?? '',
        description: json['description'] as String? ?? '',
        position: json['position'] as int? ?? 0,
        status: json['status'] as String? ?? '',
      );

  Map<String, dynamic> toWriteJson() => {
        'course_id': courseId,
        'theme': theme,
        'description': description,
        'position': position,
      };
}

class UnitDto {
  final String id;
  final String sectionId;
  final String title;
  final int position;
  final String status;

  const UnitDto({
    required this.id,
    required this.sectionId,
    required this.title,
    required this.position,
    required this.status,
  });

  factory UnitDto.fromJson(Map<String, dynamic> json) => UnitDto(
        id: json['id'] as String? ?? '',
        sectionId: json['section_id'] as String? ?? '',
        title: json['title'] as String? ?? '',
        position: json['position'] as int? ?? 0,
        status: json['status'] as String? ?? '',
      );

  Map<String, dynamic> toWriteJson() => {
        'section_id': sectionId,
        'title': title,
        'position': position,
      };
}

class SkillDto {
  final String id;
  final String unitId;
  final String title;
  final String icon;
  final int position;
  final String status;

  const SkillDto({
    required this.id,
    required this.unitId,
    required this.title,
    required this.icon,
    required this.position,
    required this.status,
  });

  factory SkillDto.fromJson(Map<String, dynamic> json) => SkillDto(
        id: json['id'] as String? ?? '',
        unitId: json['unit_id'] as String? ?? '',
        title: json['title'] as String? ?? '',
        icon: json['icon'] as String? ?? '📜',
        position: json['position'] as int? ?? 0,
        status: json['status'] as String? ?? '',
      );

  Map<String, dynamic> toWriteJson() => {
        'unit_id': unitId,
        'title': title,
        'icon': icon,
        'position': position,
      };
}

class ChallengeDto {
  final String id;
  final String skillId;
  final String type;
  final String difficulty;
  final List<String> tags;
  final int level;
  final int lessonCount;
  final String prompt;
  final String body;
  final dynamic payload;
  final dynamic options;
  final dynamic answers;
  final String explanation;
  final int position;
  final String status;

  const ChallengeDto({
    required this.id,
    required this.skillId,
    required this.type,
    required this.difficulty,
    required this.tags,
    required this.level,
    required this.lessonCount,
    required this.prompt,
    required this.body,
    required this.payload,
    required this.options,
    required this.answers,
    required this.explanation,
    required this.position,
    required this.status,
  });

  factory ChallengeDto.fromJson(Map<String, dynamic> json) => ChallengeDto(
        id: json['id'] as String? ?? '',
        skillId: json['skill_id'] as String? ?? '',
        type: json['type'] as String? ?? '',
        difficulty: json['difficulty'] as String? ?? 'easy',
        tags: _stringList(json['tags']),
        level: json['level'] as int? ?? 1,
        lessonCount: json['lesson_count'] as int? ?? 1,
        prompt: json['prompt'] as String? ?? '',
        body: json['body'] as String? ?? '',
        payload: json['payload'],
        options: json['options'],
        answers: json['answers'],
        explanation: json['explanation'] as String? ?? '',
        position: json['position'] as int? ?? 0,
        status: json['status'] as String? ?? '',
      );

  Map<String, dynamic> toWriteJson() => {
        'skill_id': skillId,
        'type': type,
        'difficulty': difficulty,
        'tags': tags,
        'level': level,
        'lesson_count': lessonCount,
        'prompt': prompt,
        'body': body,
        'payload': payload ?? {},
        'options': options ?? [],
        'answers': answers ?? [],
        'explanation': explanation,
        'position': position,
      };

  String prettyJson(dynamic value) {
    const encoder = JsonEncoder.withIndent('  ');
    return encoder.convert(value ?? {});
  }
}

class ContentApi {
  final ApiClient _client;

  const ContentApi(this._client);

  Future<List<CourseDto>> listCourses() async {
    final json = await _client.get('/courses');
    return _list(json, CourseDto.fromJson);
  }

  Future<List<SectionDto>> listSections(String courseId) async {
    final json = await _client.get('/courses/$courseId/sections');
    return _list(json, SectionDto.fromJson);
  }

  Future<List<UnitDto>> listUnits(String sectionId) async {
    final json = await _client.get('/sections/$sectionId/units');
    return _list(json, UnitDto.fromJson);
  }

  Future<List<SkillDto>> listSkills(String unitId) async {
    final json = await _client.get('/units/$unitId/skills');
    return _list(json, SkillDto.fromJson);
  }

  Future<List<ChallengeDto>> listChallenges(String skillId) async {
    final json = await _client.get('/skills/$skillId/challenges');
    return _list(json, ChallengeDto.fromJson);
  }
}

class AuthoringContentApi {
  final ApiClient _client;

  const AuthoringContentApi(this._client);

  Future<List<CourseDto>> listCourses() async {
    final json = await _client.get('/editor/content/courses', auth: true);
    return _list(json, CourseDto.fromJson);
  }

  Future<List<SectionDto>> listSections(String courseId) async {
    final json = await _client.get('/editor/content/courses/$courseId/sections', auth: true);
    return _list(json, SectionDto.fromJson);
  }

  Future<List<UnitDto>> listUnits(String sectionId) async {
    final json = await _client.get('/editor/content/sections/$sectionId/units', auth: true);
    return _list(json, UnitDto.fromJson);
  }

  Future<List<SkillDto>> listSkills(String unitId) async {
    final json = await _client.get('/editor/content/units/$unitId/skills', auth: true);
    return _list(json, SkillDto.fromJson);
  }

  Future<List<ChallengeDto>> listChallenges(String skillId) async {
    final json = await _client.get('/editor/content/skills/$skillId/challenges', auth: true);
    return _list(json, ChallengeDto.fromJson);
  }

  Future<CourseDto> saveCourse(CourseDto course) async {
    final json = course.id.isEmpty
        ? await _client.post('/editor/content/courses', body: course.toWriteJson(), auth: true)
        : await _client.patch('/editor/content/courses/${course.id}', body: course.toWriteJson(), auth: true);
    return CourseDto.fromJson(json as Map<String, dynamic>);
  }

  Future<SectionDto> saveSection(SectionDto section) async {
    final json = section.id.isEmpty
        ? await _client.post('/editor/content/sections', body: section.toWriteJson(), auth: true)
        : await _client.patch('/editor/content/sections/${section.id}', body: section.toWriteJson(), auth: true);
    return SectionDto.fromJson(json as Map<String, dynamic>);
  }

  Future<UnitDto> saveUnit(UnitDto unit) async {
    final json = unit.id.isEmpty
        ? await _client.post('/editor/content/units', body: unit.toWriteJson(), auth: true)
        : await _client.patch('/editor/content/units/${unit.id}', body: unit.toWriteJson(), auth: true);
    return UnitDto.fromJson(json as Map<String, dynamic>);
  }

  Future<SkillDto> saveSkill(SkillDto skill) async {
    final json = skill.id.isEmpty
        ? await _client.post('/editor/content/skills', body: skill.toWriteJson(), auth: true)
        : await _client.patch('/editor/content/skills/${skill.id}', body: skill.toWriteJson(), auth: true);
    return SkillDto.fromJson(json as Map<String, dynamic>);
  }

  Future<ChallengeDto> saveChallenge(ChallengeDto challenge) async {
    final json = challenge.id.isEmpty
        ? await _client.post('/editor/content/challenges', body: challenge.toWriteJson(), auth: true)
        : await _client.patch('/editor/content/challenges/${challenge.id}', body: challenge.toWriteJson(), auth: true);
    return ChallengeDto.fromJson(json as Map<String, dynamic>);
  }

  Future<void> publish(String entity, String id) async {
    await _client.post('/editor/content/$entity/$id/publish', auth: true);
  }

  Future<void> archive(String entity, String id) async {
    await _client.post('/editor/content/$entity/$id/archive', auth: true);
  }
}

List<T> _list<T>(dynamic json, T Function(Map<String, dynamic>) fromJson) {
  final items = json is List ? json : const [];
  return items.map((item) => fromJson(item as Map<String, dynamic>)).toList();
}

List<String> _stringList(dynamic raw) {
  if (raw is List) {
    return raw.map((item) => item.toString()).where((item) => item.trim().isNotEmpty).toList();
  }
  return const [];
}

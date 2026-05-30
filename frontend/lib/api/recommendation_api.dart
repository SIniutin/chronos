import 'api_client.dart';

class RecommendationDto {
  final String type;
  final String courseId;
  final String? unitId;
  final String? skillId;
  final String reason;

  const RecommendationDto({
    required this.type,
    required this.courseId,
    this.unitId,
    this.skillId,
    required this.reason,
  });

  factory RecommendationDto.fromJson(Map<String, dynamic> json) => RecommendationDto(
        type: json['type'] as String? ?? '',
        courseId: json['course_id'] as String? ?? '',
        unitId: json['unit_id'] as String?,
        skillId: json['skill_id'] as String?,
        reason: json['reason'] as String? ?? '',
      );
}

class RecommendationApi {
  final ApiClient _client;

  const RecommendationApi(this._client);

  Future<RecommendationDto?> getNext(String courseId) async {
    final encoded = Uri.encodeQueryComponent(courseId);
    final json = await _client.get('/recommendations/next?course_id=$encoded', auth: true);
    if (json == null) return null;
    return RecommendationDto.fromJson(json as Map<String, dynamic>);
  }
}

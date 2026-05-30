import 'api_client.dart';

class GamificationProfileDto {
  final int totalXp;
  final int level;
  final int currentStreak;
  final int longestStreak;
  final List<AchievementDto> achievements;

  const GamificationProfileDto({
    required this.totalXp,
    required this.level,
    required this.currentStreak,
    required this.longestStreak,
    required this.achievements,
  });

  factory GamificationProfileDto.fromJson(Map<String, dynamic> json) => GamificationProfileDto(
        totalXp: json['total_xp'] as int? ?? 0,
        level: json['level'] as int? ?? 1,
        currentStreak: json['current_streak'] as int? ?? 0,
        longestStreak: json['longest_streak'] as int? ?? 0,
        achievements: (json['achievements'] as List? ?? const [])
            .map((item) => AchievementDto.fromJson(item as Map<String, dynamic>))
            .toList(),
      );
}

class AchievementDto {
  final String code;
  final String title;
  final String description;
  final int xpReward;

  const AchievementDto({
    required this.code,
    required this.title,
    required this.description,
    required this.xpReward,
  });

  factory AchievementDto.fromJson(Map<String, dynamic> json) => AchievementDto(
        code: json['code'] as String? ?? '',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        xpReward: json['xp_reward'] as int? ?? 0,
      );
}

class GamificationApi {
  final ApiClient _client;

  const GamificationApi(this._client);

  Future<GamificationProfileDto> getProfile() async {
    final json = await _client.get('/gamification/profile', auth: true);
    return GamificationProfileDto.fromJson(json as Map<String, dynamic>);
  }
}

import 'dart:convert';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:latlong2/latlong.dart';

import '../api/api_client.dart';
import '../api/auth_api.dart';
import '../api/content_api.dart';
import '../api/media_api.dart';
import '../api/progress_api.dart';
import '../state/session_controller.dart';
import '../theme/app_theme.dart';
import '../widgets/responsive_text.dart';

class AdminPanelPage extends StatefulWidget {
  const AdminPanelPage({super.key});

  @override
  State<AdminPanelPage> createState() => _AdminPanelPageState();
}

class _AdminPanelPageState extends State<AdminPanelPage> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    AppTheme.currentMode = session.themeMode;
    final user = session.currentUser;
    final tabs = [
      const _PanelTab('Контент', Icons.account_tree_outlined),
      const _PanelTab('Публикация', Icons.verified_outlined),
      if (user?.isAdmin == true)
        const _PanelTab('Пользователи', Icons.admin_panel_settings_outlined),
    ];

    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: const Text('Панель управления'),
        backgroundColor: AppTheme.primary,
      ),
      body: Column(
        children: [
          SizedBox(
            height: 58,
            child: ListView.separated(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              scrollDirection: Axis.horizontal,
              itemBuilder: (_, i) => ChoiceChip(
                selected: _tab == i,
                avatar: Icon(tabs[i].icon,
                    size: 16,
                    color: _tab == i ? AppTheme.onAccent : AppTheme.accent),
                label: ResponsiveText(tabs[i].label),
                onSelected: (_) => setState(() => _tab = i),
                selectedColor: AppTheme.accent,
                backgroundColor: AppTheme.surface,
                labelStyle: GoogleFonts.lato(
                    color:
                        _tab == i ? AppTheme.onAccent : AppTheme.textPrimary),
              ),
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemCount: tabs.length,
            ),
          ),
          Expanded(
            child: switch (tabs[_tab].label) {
              'Контент' => const _ContentTreeEditor(mode: _ContentMode.edit),
              'Публикация' =>
                const _ContentTreeEditor(mode: _ContentMode.review),
              _ => const _UsersRolePanel(),
            },
          ),
        ],
      ),
    );
  }
}

class _PanelTab {
  final String label;
  final IconData icon;

  const _PanelTab(this.label, this.icon);
}

enum _ContentMode { edit, review }

String _statusLabel(String status) => switch (status) {
      'draft' => 'черновик',
      'updating' => 'на правке',
      'published' => 'опубликовано',
      'archived' => 'архив',
      _ => status,
    };

String _roleLabel(String role) => switch (role) {
      'student' => 'Студент',
      'content_editor' => 'Редактор контента',
      'content_reviewer' => 'Проверяющий',
      'admin' => 'Администратор',
      _ => role,
    };

String _difficultyLabel(String difficulty) => switch (difficulty) {
      'easy' => 'Легкая',
      'medium' => 'Средняя',
      'hard' => 'Сложная',
      _ => difficulty,
    };

String _challengeTypeLabel(String type) => switch (type) {
      'theory' => 'Теория',
      'single_choice' => 'Один правильный ответ',
      'multiple_choice' => 'Несколько правильных ответов',
      'timeline' => 'Хронология событий',
      'match_pairs' => 'Соединить пары',
      'image_question' => 'Вопрос по изображению',
      'match_image' => 'Сопоставить изображения',
      'match_photos' => 'Сопоставить фотографии',
      'quote_question' => 'Вопрос по цитате',
      'true_false' => 'Верно / неверно',
      'fill_in_blank' => 'Заполнить пропуск',
      'map_point' => 'Точка на карте',
      'map_area' => 'Обвести область',
      _ => type,
    };

bool _isRealMediaUrl(String url) {
  final value = url.trim();
  return value.startsWith('http://') ||
      value.startsWith('https://') ||
      value.startsWith('/media/');
}

class _ContentTreeEditor extends StatefulWidget {
  final _ContentMode mode;

  const _ContentTreeEditor({required this.mode});

  @override
  State<_ContentTreeEditor> createState() => _ContentTreeEditorState();
}

class _ContentTreeEditorState extends State<_ContentTreeEditor> {
  late AuthoringContentApi _api;
  List<CourseDto> _courses = [];
  List<SectionDto> _sections = [];
  List<UnitDto> _units = [];
  List<SkillDto> _skills = [];
  List<ChallengeDto> _challenges = [];
  final Map<String, _SkillChallengeStats> _skillStats = {};
  CourseDto? _course;
  SectionDto? _section;
  UnitDto? _unit;
  SkillDto? _skill;
  bool _loading = true;
  String? _error;

  bool get _canEdit =>
      SessionScope.of(context).currentUser?.canEditContent == true &&
      widget.mode == _ContentMode.edit;
  bool get _canPublish =>
      SessionScope.of(context).currentUser?.canReviewContent == true;
  bool get _canArchive => SessionScope.of(context).currentUser?.isAdmin == true;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _api = AuthoringContentApi(SessionScope.of(context).client);
    _loadCourses();
  }

  Future<void> _guard(Future<void> Function() action) async {
    setState(() => _error = null);
    try {
      await action();
    } on ApiException catch (e) {
      setState(() {
        _loading = false;
        _error = e.message;
      });
    } catch (_) {
      setState(() {
        _loading = false;
        _error = 'Ошибка запроса';
      });
    }
  }

  Future<void> _loadCourses() => _guard(() async {
        final courses = await _api.listCourses();
        setState(() {
          _courses = courses;
          _loading = false;
        });
      });

  Future<void> _selectCourse(CourseDto course) => _guard(() async {
        final sections = await _api.listSections(course.id);
        setState(() {
          _course = course;
          _section = null;
          _unit = null;
          _skill = null;
          _sections = sections;
          _units = [];
          _skills = [];
          _skillStats.clear();
          _challenges = [];
        });
      });

  Future<void> _selectSection(SectionDto section) => _guard(() async {
        final units = await _api.listUnits(section.id);
        setState(() {
          _section = section;
          _unit = null;
          _skill = null;
          _units = units;
          _skills = [];
          _skillStats.clear();
          _challenges = [];
        });
      });

  Future<void> _selectUnit(UnitDto unit) => _guard(() async {
        final skills = await _api.listSkills(unit.id);
        final stats = <String, _SkillChallengeStats>{};
        for (final skill in skills) {
          final challenges = await _api.listChallenges(skill.id);
          stats[skill.id] = _SkillChallengeStats.fromChallenges(challenges);
        }
        setState(() {
          _unit = unit;
          _skill = null;
          _skills = skills;
          _skillStats
            ..clear()
            ..addAll(stats);
          _challenges = [];
        });
      });

  Future<void> _selectSkill(SkillDto skill) => _guard(() async {
        final challenges = await _api.listChallenges(skill.id);
        setState(() {
          _skill = skill;
          _challenges = challenges;
        });
      });

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Center(
          child: CircularProgressIndicator(color: AppTheme.accent));
    }
    return RefreshIndicator(
      onRefresh: _loadCourses,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (_error != null) _ErrorBanner(_error!),
          _HeaderRow(
            title: 'Курсы',
            canAdd: _canEdit,
            onAdd: () => _editCourse(null),
          ),
          ..._courses.map((course) => _EntityTile(
                title: course.title,
                subtitle: 'Языки: ${course.sourceLang} → ${course.targetLang}',
                status: course.status,
                selected: _course?.id == course.id,
                onTap: () => _selectCourse(course),
                onEdit: _canEdit ? () => _editCourse(course) : null,
                onPublish: _canPublish
                    ? () => _transition('courses', course.id, publish: true)
                    : null,
                onArchive: _canArchive
                    ? () => _transition('courses', course.id, publish: false)
                    : null,
              )),
          if (_course != null) ...[
            _HeaderRow(
                title: 'Разделы',
                canAdd: _canEdit,
                onAdd: () => _editSection(null)),
            ..._sections.map((section) => _EntityTile(
                  title: section.theme,
                  subtitle: section.description,
                  status: section.status,
                  selected: _section?.id == section.id,
                  onTap: () => _selectSection(section),
                  onEdit: _canEdit ? () => _editSection(section) : null,
                  onPublish: _canPublish
                      ? () => _transition('sections', section.id, publish: true)
                      : null,
                  onArchive: _canArchive
                      ? () =>
                          _transition('sections', section.id, publish: false)
                      : null,
                )),
          ],
          if (_section != null) ...[
            _HeaderRow(
                title: 'Уроки', canAdd: _canEdit, onAdd: () => _editUnit(null)),
            ..._units.map((unit) => _EntityTile(
                  title: unit.title,
                  subtitle: 'Порядок показа: ${unit.position}',
                  status: unit.status,
                  selected: _unit?.id == unit.id,
                  onTap: () => _selectUnit(unit),
                  onEdit: _canEdit ? () => _editUnit(unit) : null,
                  onPublish: _canPublish
                      ? () => _transition('units', unit.id, publish: true)
                      : null,
                  onArchive: _canArchive
                      ? () => _transition('units', unit.id, publish: false)
                      : null,
                )),
          ],
          if (_unit != null) ...[
            _HeaderRow(
                title: 'Навыки',
                canAdd: _canEdit,
                onAdd: () => _editSkill(null)),
            ..._skills.map((skill) {
              final stats = _skillStats[skill.id];
              return _EntityTile(
                title: '${skill.icon} ${skill.title}',
                subtitle: [
                  'Порядок показа: ${skill.position}',
                  if (stats != null) stats.summary,
                  if (stats != null && stats.warning.isNotEmpty) stats.warning,
                ].join(' · '),
                status: skill.status,
                selected: _skill?.id == skill.id,
                onTap: () => _selectSkill(skill),
                onEdit: _canEdit ? () => _editSkill(skill) : null,
                onPublish: _canPublish
                    ? () => _transition('skills', skill.id, publish: true)
                    : null,
                onArchive: _canArchive
                    ? () => _transition('skills', skill.id, publish: false)
                    : null,
              );
            }),
          ],
          if (_skill != null) ...[
            _HeaderRow(
                title: 'Вопросы и теория',
                canAdd: _canEdit,
                onAdd: () => _editChallenge(null)),
            ..._challenges.map((challenge) => _EntityTile(
                  title: challenge.prompt,
                  subtitle:
                      '${_challengeTypeLabel(challenge.type)} · ${_difficultyLabel(challenge.difficulty)}${challenge.tags.contains('needs_review') ? ' · needs_review' : ''}',
                  status: challenge.status,
                  selected: false,
                  onTap: _canEdit ? () => _editChallenge(challenge) : () {},
                  onEdit: _canEdit ? () => _editChallenge(challenge) : null,
                  onPublish: _canPublish
                      ? () =>
                          _transition('challenges', challenge.id, publish: true)
                      : null,
                  onArchive: _canArchive
                      ? () => _transition('challenges', challenge.id,
                          publish: false)
                      : null,
                )),
          ],
        ],
      ),
    );
  }

  Future<void> _transition(String entity, String id, {required bool publish}) =>
      _guard(() async {
        if (publish) {
          await _api.publish(entity, id);
        } else {
          await _api.archive(entity, id);
        }
        await _reloadCurrent();
      });

  Future<void> _reloadCurrent() async {
    await _loadCourses();
    final course = _course;
    final section = _section;
    final unit = _unit;
    final skill = _skill;
    if (course != null) await _selectCourse(course);
    if (section != null) await _selectSection(section);
    if (unit != null) await _selectUnit(unit);
    if (skill != null) await _selectSkill(skill);
  }

  Future<void> _editCourse(CourseDto? course) async {
    final saved = await showDialog<CourseDto>(
      context: context,
      builder: (_) => _CourseDialog(course: course),
    );
    if (saved == null) return;
    await _guard(() async {
      await _api.saveCourse(saved);
      await _loadCourses();
    });
  }

  Future<void> _editSection(SectionDto? section) async {
    final course = _course;
    if (course == null) return;
    final saved = await showDialog<SectionDto>(
      context: context,
      builder: (_) => _SectionDialog(courseId: course.id, section: section),
    );
    if (saved == null) return;
    await _guard(() async {
      await _api.saveSection(saved);
      await _selectCourse(course);
    });
  }

  Future<void> _editUnit(UnitDto? unit) async {
    final section = _section;
    if (section == null) return;
    final saved = await showDialog<UnitDto>(
      context: context,
      builder: (_) => _UnitDialog(sectionId: section.id, unit: unit),
    );
    if (saved == null) return;
    await _guard(() async {
      await _api.saveUnit(saved);
      await _selectSection(section);
    });
  }

  Future<void> _editSkill(SkillDto? skill) async {
    final unit = _unit;
    if (unit == null) return;
    final saved = await showDialog<SkillDto>(
      context: context,
      builder: (_) => _SkillDialog(unitId: unit.id, skill: skill),
    );
    if (saved == null) return;
    await _guard(() async {
      await _api.saveSkill(saved);
      await _selectUnit(unit);
    });
  }

  Future<void> _editChallenge(ChallengeDto? challenge) async {
    final skill = _skill;
    if (skill == null || !_canEdit && challenge == null) return;
    final saved = await showDialog<ChallengeDto>(
      context: context,
      builder: (_) => _ChallengeDialog(skillId: skill.id, challenge: challenge),
    );
    if (saved == null) return;
    await _guard(() async {
      await _api.saveChallenge(saved);
      await _selectSkill(skill);
    });
  }
}

class _HeaderRow extends StatelessWidget {
  final String title;
  final bool canAdd;
  final VoidCallback onAdd;

  const _HeaderRow(
      {required this.title, required this.canAdd, required this.onAdd});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 18, bottom: 8),
      child: Row(
        children: [
          Expanded(
            child: Text(title,
                style: GoogleFonts.playfairDisplay(
                    color: AppTheme.textPrimary,
                    fontSize: 20,
                    fontWeight: FontWeight.bold)),
          ),
          if (canAdd)
            IconButton(
              tooltip: 'Добавить',
              onPressed: onAdd,
              icon:
                  const Icon(Icons.add_circle_outline, color: AppTheme.accent),
            ),
        ],
      ),
    );
  }
}

class _SkillChallengeStats {
  final int total;
  final int published;
  final int draft;
  final int placeholder;
  final int interactive;
  final int learnerVisible;
  final int learnerVisibleInteractive;
  final bool hasTheory;
  final bool hasPublishedPlaceholder;
  final bool hasUnsafePublishedPhotos;

  const _SkillChallengeStats({
    required this.total,
    required this.published,
    required this.draft,
    required this.placeholder,
    required this.interactive,
    required this.learnerVisible,
    required this.learnerVisibleInteractive,
    required this.hasTheory,
    required this.hasPublishedPlaceholder,
    required this.hasUnsafePublishedPhotos,
  });

  factory _SkillChallengeStats.fromChallenges(List<ChallengeDto> challenges) {
    var published = 0;
    var draft = 0;
    var placeholder = 0;
    var interactive = 0;
    var learnerVisible = 0;
    var learnerVisibleInteractive = 0;
    var hasTheory = false;
    var hasPublishedPlaceholder = false;
    var hasUnsafePublishedPhotos = false;
    for (final challenge in challenges) {
      final isPublished = challenge.status == 'published';
      final isDraft = challenge.status == 'draft';
      final isPlaceholder = challenge.tags.contains('placeholder') ||
          challenge.tags.contains('needs_review');
      final isInteractive = _isInteractiveType(challenge.type);
      final unsafePhotos = challenge.type == 'match_photos' &&
          _matchPhotosHasEmptyImageUrl(challenge.options);
      if (isPublished) published++;
      if (isDraft) draft++;
      if (isPlaceholder) placeholder++;
      if (isInteractive) interactive++;
      if (challenge.type == 'theory') hasTheory = true;
      if (isPublished && isPlaceholder) hasPublishedPlaceholder = true;
      if (isPublished && unsafePhotos) hasUnsafePublishedPhotos = true;
      if (isPublished && !isPlaceholder && !unsafePhotos) {
        learnerVisible++;
        if (isInteractive) learnerVisibleInteractive++;
      }
    }
    return _SkillChallengeStats(
      total: challenges.length,
      published: published,
      draft: draft,
      placeholder: placeholder,
      interactive: interactive,
      learnerVisible: learnerVisible,
      learnerVisibleInteractive: learnerVisibleInteractive,
      hasTheory: hasTheory,
      hasPublishedPlaceholder: hasPublishedPlaceholder,
      hasUnsafePublishedPhotos: hasUnsafePublishedPhotos,
    );
  }

  String get summary =>
      'Всего: $total, published: $published, draft: $draft, placeholder: $placeholder, interactive: $interactive';

  String get warning {
    final warnings = <String>[];
    if (learnerVisible > 7) warnings.add('>7 learner-visible');
    if (learnerVisibleInteractive > 1) warnings.add('>1 interactive');
    if (!hasTheory) warnings.add('нет theory');
    if (hasPublishedPlaceholder) warnings.add('published placeholder');
    if (hasUnsafePublishedPhotos) warnings.add('empty photo URL');
    return warnings.isEmpty ? '' : '⚠ ${warnings.join(', ')}';
  }

  static bool _isInteractiveType(String type) =>
      type == 'map_point' || type == 'map_area' || type == 'match_photos';

  static bool _matchPhotosHasEmptyImageUrl(dynamic options) {
    if (options is! Map || options['photos'] is! List) return true;
    final photos = options['photos'] as List;
    if (photos.isEmpty) return true;
    return photos.any((photo) =>
        photo is! Map ||
        !_isRealMediaUrl('${photo['image_url'] ?? photo['imageUrl'] ?? ''}'));
  }
}

class _EntityTile extends StatelessWidget {
  final String title;
  final String subtitle;
  final String status;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback? onEdit;
  final VoidCallback? onPublish;
  final VoidCallback? onArchive;

  const _EntityTile({
    required this.title,
    required this.subtitle,
    required this.status,
    required this.selected,
    required this.onTap,
    this.onEdit,
    this.onPublish,
    this.onArchive,
  });

  @override
  Widget build(BuildContext context) {
    final actions = <PopupMenuEntry<String>>[
      if (onEdit != null)
        const PopupMenuItem(value: 'edit', child: Text('Редактировать')),
      if (onPublish != null)
        const PopupMenuItem(value: 'publish', child: Text('Опубликовать')),
      if (onArchive != null)
        const PopupMenuItem(value: 'archive', child: Text('В архив')),
    ];
    return Card(
      color: selected ? AppTheme.cardBg : AppTheme.surface,
      shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side:
              BorderSide(color: selected ? AppTheme.accent : AppTheme.cardBg)),
      child: ListTile(
        onTap: onTap,
        title: Text(
          title.isEmpty ? 'Без названия' : title,
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
          style: GoogleFonts.lato(
              color: AppTheme.textPrimary, fontWeight: FontWeight.bold),
        ),
        subtitle: Text(
          '$subtitle · ${_statusLabel(status)}',
          maxLines: 4,
          overflow: TextOverflow.ellipsis,
          style: GoogleFonts.lato(color: AppTheme.textSecondary),
        ),
        trailing: actions.isEmpty
            ? null
            : PopupMenuButton<String>(
                icon: Icon(Icons.more_vert, color: AppTheme.textSecondary),
                color: AppTheme.surface,
                tooltip: 'Действия',
                onSelected: (value) {
                  switch (value) {
                    case 'edit':
                      onEdit?.call();
                      break;
                    case 'publish':
                      onPublish?.call();
                      break;
                    case 'archive':
                      onArchive?.call();
                      break;
                  }
                },
                itemBuilder: (_) => actions,
              ),
      ),
    );
  }
}

class _UsersRolePanel extends StatefulWidget {
  const _UsersRolePanel();

  @override
  State<_UsersRolePanel> createState() => _UsersRolePanelState();
}

class _UsersRolePanelState extends State<_UsersRolePanel> {
  final _identity = TextEditingController();
  AppUser? _found;
  String _role = 'student';
  String? _message;

  @override
  void dispose() {
    _identity.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final session = SessionScope.of(context);
    final auth = AuthApi(session.client);
    final progress = ProgressApi(session.client);
    final currentUser = session.currentUser;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text('Тестовый прогресс',
            style: GoogleFonts.playfairDisplay(
                color: AppTheme.textPrimary,
                fontSize: 22,
                fontWeight: FontWeight.bold)),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: AppTheme.surface,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: AppTheme.cardBg),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              ResponsiveText(
                'Быстро отметить весь курс пройденным для вашего аккаунта.',
                style: GoogleFonts.lato(
                    color: AppTheme.textSecondary, fontSize: 14),
              ),
              const SizedBox(height: 12),
              ElevatedButton.icon(
                onPressed: currentUser?.id == null || currentUser!.id!.isEmpty
                    ? null
                    : () async {
                        try {
                          await progress.completeAllForUser(currentUser.id!);
                          if (!mounted) return;
                          setState(() => _message =
                              'Весь курс открыт для вашего аккаунта');
                        } on ApiException catch (e) {
                          if (!mounted) return;
                          setState(() => _message = e.message);
                        }
                      },
                icon: const Icon(Icons.workspace_premium_outlined),
                label: const ButtonLabel('Открыть весь курс себе'),
              ),
            ],
          ),
        ),
        if (_message != null) ...[
          const SizedBox(height: 14),
          ResponsiveText(_message!,
              style: GoogleFonts.lato(
                  color: AppTheme.accent, fontWeight: FontWeight.bold)),
        ],
        const SizedBox(height: 24),
        Text('Смена роли',
            style: GoogleFonts.playfairDisplay(
                color: AppTheme.textPrimary,
                fontSize: 22,
                fontWeight: FontWeight.bold)),
        const SizedBox(height: 12),
        _TextInput(controller: _identity, label: 'Email или логин'),
        const SizedBox(height: 12),
        ElevatedButton.icon(
          onPressed: () async {
            try {
              final user = await auth.findUser(_identity.text.trim());
              setState(() {
                _found = user;
                _role = user.role;
                _message = null;
              });
            } on ApiException catch (e) {
              setState(() => _message = e.message);
            }
          },
          icon: const Icon(Icons.search),
          label: const ButtonLabel('Найти пользователя'),
        ),
        if (_found != null) ...[
          const SizedBox(height: 18),
          _InfoBox('${_found!.login} · ${_found!.email}',
              'Текущая роль: ${_roleLabel(_found!.role)}'),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            initialValue: _role,
            dropdownColor: AppTheme.surface,
            decoration: _inputDecoration('Новая роль'),
            items: const [
              'student',
              'content_editor',
              'content_reviewer',
              'admin'
            ]
                .map((role) => DropdownMenuItem(
                    value: role, child: ResponsiveText(_roleLabel(role))))
                .toList(),
            onChanged: (value) => setState(() => _role = value ?? _role),
          ),
          const SizedBox(height: 12),
          ElevatedButton.icon(
            onPressed: () async {
              try {
                final updated = await auth.changeRole(_found!.id!, _role);
                setState(() {
                  _found = updated;
                  _message = 'Роль обновлена';
                });
              } on ApiException catch (e) {
                setState(() => _message = e.message);
              }
            },
            icon: const Icon(Icons.save_outlined),
            label: const ButtonLabel('Сохранить роль'),
          ),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: () async {
              final userId = _found!.id;
              if (userId == null || userId.isEmpty) return;
              final confirmed = await showDialog<bool>(
                context: context,
                builder: (_) => AlertDialog(
                  title: const Text('Отметить всё пройденным?'),
                  content: Text(
                      'Пользователь ${_found!.login} получит completed-прогресс по всем опубликованным урокам.'),
                  actions: [
                    TextButton(
                        onPressed: () => Navigator.pop(context, false),
                        child: const Text('Отмена')),
                    ElevatedButton(
                        onPressed: () => Navigator.pop(context, true),
                        child: const Text('Подтвердить')),
                  ],
                ),
              );
              if (confirmed != true) return;
              try {
                await progress.completeAllForUser(userId);
                setState(() =>
                    _message = 'Все опубликованные уроки отмечены пройденными');
              } on ApiException catch (e) {
                setState(() => _message = e.message);
              }
            },
            icon: const Icon(Icons.done_all_outlined),
            label: const ButtonLabel('Отметить всё пройденным'),
          ),
        ],
      ],
    );
  }
}

class _CourseDialog extends StatelessWidget {
  final CourseDto? course;
  final _source = TextEditingController();
  final _target = TextEditingController();
  final _title = TextEditingController();

  _CourseDialog({required this.course}) {
    _source.text = course?.sourceLang ?? 'ru';
    _target.text = course?.targetLang ?? 'ru';
    _title.text = course?.title ?? '';
  }

  @override
  Widget build(BuildContext context) {
    return _FormDialog(
      title: course == null ? 'Новый курс' : 'Редактировать курс',
      fields: [
        _TextInput(controller: _source, label: 'Язык материалов'),
        _TextInput(controller: _target, label: 'Язык обучения'),
        _TextInput(controller: _title, label: 'Название курса'),
      ],
      onSave: () => Navigator.pop(
          context,
          CourseDto(
              id: course?.id ?? '',
              sourceLang: _source.text,
              targetLang: _target.text,
              title: _title.text,
              status: course?.status ?? 'draft')),
    );
  }
}

class _SectionDialog extends StatelessWidget {
  final String courseId;
  final SectionDto? section;
  final _theme = TextEditingController();
  final _description = TextEditingController();
  final _position = TextEditingController();

  _SectionDialog({required this.courseId, required this.section}) {
    _theme.text = section?.theme ?? '';
    _description.text = section?.description ?? '';
    _position.text = '${section?.position ?? 1}';
  }

  @override
  Widget build(BuildContext context) {
    return _FormDialog(
      title: section == null ? 'Новый раздел' : 'Редактировать раздел',
      fields: [
        _TextInput(controller: _theme, label: 'Название раздела'),
        _TextInput(
            controller: _description, label: 'Краткое описание', maxLines: 3),
        _TextInput(
            controller: _position,
            label: 'Порядок показа',
            keyboardType: TextInputType.number),
      ],
      onSave: () => Navigator.pop(
          context,
          SectionDto(
              id: section?.id ?? '',
              courseId: courseId,
              theme: _theme.text,
              description: _description.text,
              position: int.tryParse(_position.text) ?? 1,
              status: section?.status ?? 'draft')),
    );
  }
}

class _UnitDialog extends StatelessWidget {
  final String sectionId;
  final UnitDto? unit;
  final _title = TextEditingController();
  final _position = TextEditingController();

  _UnitDialog({required this.sectionId, required this.unit}) {
    _title.text = unit?.title ?? '';
    _position.text = '${unit?.position ?? 1}';
  }

  @override
  Widget build(BuildContext context) {
    return _FormDialog(
      title: unit == null ? 'Новый урок' : 'Редактировать урок',
      fields: [
        _TextInput(controller: _title, label: 'Название урока'),
        _TextInput(
            controller: _position,
            label: 'Порядок показа',
            keyboardType: TextInputType.number),
      ],
      onSave: () => Navigator.pop(
          context,
          UnitDto(
              id: unit?.id ?? '',
              sectionId: sectionId,
              title: _title.text,
              position: int.tryParse(_position.text) ?? 1,
              status: unit?.status ?? 'draft')),
    );
  }
}

class _SkillDialog extends StatelessWidget {
  final String unitId;
  final SkillDto? skill;
  final _title = TextEditingController();
  final _icon = TextEditingController();
  final _position = TextEditingController();

  _SkillDialog({required this.unitId, required this.skill}) {
    _title.text = skill?.title ?? '';
    _icon.text = skill?.icon ?? '📜';
    _position.text = '${skill?.position ?? 1}';
  }

  @override
  Widget build(BuildContext context) {
    return _FormDialog(
      title: skill == null ? 'Новый навык' : 'Редактировать навык',
      fields: [
        _TextInput(controller: _title, label: 'Название навыка'),
        _TextInput(controller: _icon, label: 'Иконка'),
        _TextInput(
            controller: _position,
            label: 'Порядок показа',
            keyboardType: TextInputType.number),
      ],
      onSave: () => Navigator.pop(
          context,
          SkillDto(
              id: skill?.id ?? '',
              unitId: unitId,
              title: _title.text,
              icon: _icon.text,
              position: int.tryParse(_position.text) ?? 1,
              status: skill?.status ?? 'draft')),
    );
  }
}

class _ChallengeDialog extends StatefulWidget {
  final String skillId;
  final ChallengeDto? challenge;

  const _ChallengeDialog({required this.skillId, required this.challenge});

  @override
  State<_ChallengeDialog> createState() => _ChallengeDialogState();
}

class _ChallengeDialogState extends State<_ChallengeDialog> {
  static const _types = [
    'theory',
    'single_choice',
    'multiple_choice',
    'timeline',
    'match_pairs',
    'image_question',
    'match_image',
    'match_photos',
    'quote_question',
    'true_false',
    'fill_in_blank',
    'map_point',
    'map_area',
  ];
  static const _difficulties = ['easy', 'medium', 'hard'];
  static const _defaultTileUrl =
      'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
  static const _defaultMapAttribution = '© OpenStreetMap contributors';
  static const _defaultMapCenter = LatLng(55.7558, 37.6173);

  final _type = TextEditingController();
  final _difficulty = TextEditingController();
  final _tags = TextEditingController();
  final _level = TextEditingController();
  final _lessonCount = TextEditingController();
  final _prompt = TextEditingController();
  final _body = TextEditingController();
  final _payload = TextEditingController();
  final _options = TextEditingController();
  final _answers = TextEditingController();
  final _explanation = TextEditingController();
  final _position = TextEditingController();
  final _status = TextEditingController();
  final _mapCenterLat = TextEditingController();
  final _mapCenterLng = TextEditingController();
  final _mapZoom = TextEditingController();
  final _mapTileUrl = TextEditingController();
  final _mapAttribution = TextEditingController();
  final _mapPointRadius = TextEditingController();
  final _mapAreaCenterRadius = TextEditingController();
  final _mapAreaTolerance = TextEditingController();
  String? _error;
  bool _uploadingImage = false;
  int? _uploadingMatchPhotoIndex;
  LatLng? _mapPointAnswer;
  final List<LatLng> _mapAreaPoints = [];
  LatLng? _mapAreaCenterAnswer;
  double? _mapAreaM2;
  bool _mapAdvancedOpen = false;

  @override
  void initState() {
    super.initState();
    final challenge = widget.challenge;
    _type.text = challenge?.type ?? 'single_choice';
    _difficulty.text = challenge?.difficulty ?? 'easy';
    _tags.text = challenge?.tags.join(', ') ?? '';
    _level.text = '${challenge?.level ?? 1}';
    _lessonCount.text = '${challenge?.lessonCount ?? 1}';
    _prompt.text = challenge?.prompt ?? '';
    _body.text = challenge?.body ?? '';
    _explanation.text = challenge?.explanation ?? '';
    _position.text = '${challenge?.position ?? 1}';
    _status.text = challenge?.status ?? 'draft';
    if (challenge == null) {
      _applyTemplate(_type.text);
    } else {
      _hydrateEditorFields(challenge);
    }
  }

  @override
  void dispose() {
    _type.dispose();
    _difficulty.dispose();
    _tags.dispose();
    _level.dispose();
    _lessonCount.dispose();
    _prompt.dispose();
    _body.dispose();
    _payload.dispose();
    _options.dispose();
    _answers.dispose();
    _explanation.dispose();
    _position.dispose();
    _status.dispose();
    _mapCenterLat.dispose();
    _mapCenterLng.dispose();
    _mapZoom.dispose();
    _mapTileUrl.dispose();
    _mapAttribution.dispose();
    _mapPointRadius.dispose();
    _mapAreaCenterRadius.dispose();
    _mapAreaTolerance.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final type = _type.text;
    return _FormDialog(
      title: widget.challenge == null
          ? 'Новый вопрос или теория'
          : 'Редактировать материал',
      error: _error,
      fields: [
        DropdownButtonFormField<String>(
          initialValue: _types.contains(type) ? type : 'single_choice',
          dropdownColor: AppTheme.surface,
          decoration: _inputDecoration('Тип материала'),
          items: _types
              .map((type) => DropdownMenuItem(
                  value: type,
                  child: ResponsiveText(_challengeTypeLabel(type))))
              .toList(),
          onChanged: (value) {
            if (value == null) return;
            setState(() {
              _type.text = value;
              _applyTemplate(value);
            });
          },
        ),
        DropdownButtonFormField<String>(
          initialValue: _difficulties.contains(_difficulty.text)
              ? _difficulty.text
              : 'easy',
          dropdownColor: AppTheme.surface,
          decoration: _inputDecoration('Сложность'),
          items: _difficulties
              .map((item) => DropdownMenuItem(
                  value: item, child: ResponsiveText(_difficultyLabel(item))))
              .toList(),
          onChanged: (value) =>
              setState(() => _difficulty.text = value ?? 'easy'),
        ),
        DropdownButtonFormField<String>(
          initialValue: ['draft', 'published', 'archived', 'updating']
                  .contains(_status.text)
              ? _status.text
              : 'draft',
          dropdownColor: AppTheme.surface,
          decoration: _inputDecoration('Статус'),
          items: const [
            DropdownMenuItem(value: 'draft', child: Text('draft')),
            DropdownMenuItem(value: 'published', child: Text('published')),
            DropdownMenuItem(value: 'updating', child: Text('updating')),
            DropdownMenuItem(value: 'archived', child: Text('archived')),
          ],
          onChanged: (value) => setState(() => _status.text = value ?? 'draft'),
        ),
        _TextInput(controller: _tags, label: 'Теги через запятую'),
        _TextInput(
            controller: _level,
            label: 'Уровень внутри навыка',
            keyboardType: TextInputType.number),
        _TextInput(
            controller: _lessonCount,
            label: 'Количество шагов урока',
            keyboardType: TextInputType.number),
        _TextInput(controller: _prompt, label: _promptLabel(type), maxLines: 3),
        if (_usesBody(type))
          _TextInput(controller: _body, label: 'Текст теории', maxLines: 4),
        if (_usesImageUpload(type))
          OutlinedButton.icon(
            onPressed: _uploadingImage ? null : () => _pickAndUploadImage(type),
            icon: _uploadingImage
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.upload_file_outlined),
            label: ButtonLabel(
                _uploadingImage ? 'Загрузка...' : 'Загрузить изображение'),
          ),
        if (_isMapType(type))
          _MapAuthoringEditor(
            type: type,
            center: _mapCenter(),
            zoom: _mapZoomValue(),
            tileUrlTemplate: _mapTileUrl.text.trim().isEmpty
                ? _defaultTileUrl
                : _mapTileUrl.text.trim(),
            attribution: _mapAttribution.text.trim().isEmpty
                ? _defaultMapAttribution
                : _mapAttribution.text.trim(),
            pointAnswer: _mapPointAnswer,
            areaPoints: _mapAreaPoints,
            areaCenter: _mapAreaCenterAnswer,
            areaM2: _mapAreaM2,
            advancedOpen: _mapAdvancedOpen,
            centerLatController: _mapCenterLat,
            centerLngController: _mapCenterLng,
            zoomController: _mapZoom,
            tileUrlController: _mapTileUrl,
            attributionController: _mapAttribution,
            pointRadiusController: _mapPointRadius,
            areaCenterRadiusController: _mapAreaCenterRadius,
            areaToleranceController: _mapAreaTolerance,
            onChanged: _onMapEditorChanged,
            onPointSelected: _setMapPointAnswer,
            onAreaPointAdded: _addMapAreaPoint,
            onAreaClear: _clearMapArea,
            onAreaDone: _finalizeMapArea,
            onAdvancedChanged: (value) =>
                setState(() => _mapAdvancedOpen = value),
            onCenterFromAnswer: _centerMapFromAnswer,
            onCopyCenterToAnswer: _copyMapCenterToAnswer,
            onAreaSimplify: _simplifyMapArea,
            onResetDefaults: _resetMapEditorFromButton,
          )
        else if (_usesPayload(type))
          _TextInput(
              controller: _payload, label: _payloadLabel(type), maxLines: 5),
        if (type == 'match_photos')
          _MatchPhotosAuthoringEditor(
            lines: _nonEmptyLines(_options.text),
            published: _status.text == 'published',
            uploadingIndex: _uploadingMatchPhotoIndex,
            onUpload: (index) {
              if (!_uploadingImage) {
                _pickAndUploadImage(type, photoIndex: index);
              }
            },
            onChanged: _setMatchPhotoLines,
          )
        else if (_usesOptions(type))
          _TextInput(
              controller: _options, label: _optionsLabel(type), maxLines: 5),
        if (!_isMapType(type) && _usesAnswers(type))
          _TextInput(
              controller: _answers, label: _answersLabel(type), maxLines: 5),
        _TextInput(
            controller: _explanation,
            label: _explanationLabel(type),
            maxLines: 4),
        _TextInput(
            controller: _position,
            label: 'Порядок показа',
            keyboardType: TextInputType.number),
      ],
      onSave: () {
        try {
          _validateMapEditor();
          _validateMatchPhotosEditor();
          Navigator.pop(
            context,
            ChallengeDto(
              id: widget.challenge?.id ?? '',
              skillId: widget.skillId,
              type: _type.text,
              difficulty: _difficulty.text,
              tags: _tags.text
                  .split(',')
                  .map((tag) => tag.trim())
                  .where((tag) => tag.isNotEmpty)
                  .toList(),
              level: int.tryParse(_level.text) ?? 1,
              lessonCount: int.tryParse(_lessonCount.text) ?? 1,
              prompt: _prompt.text,
              body: _body.text,
              payload: _payloadValue(),
              options: _optionsValue(),
              answers: _answersValue(),
              explanation: _explanation.text,
              position: int.tryParse(_position.text) ?? 1,
              status: _status.text.isEmpty ? 'draft' : _status.text,
            ),
          );
        } catch (_) {
          setState(() => _error =
              'Проверьте поля с вариантами, ответами и дополнительными данными');
        }
      },
    );
  }

  void _applyTemplate(String type) {
    switch (type) {
      case 'theory':
        _payload.text = 'Факт 1\nФакт 2\n\nКраткое резюме';
        _options.clear();
        _answers.clear();
      case 'timeline':
        _payload.clear();
        _options.text =
            '1905 | Первая русская революция\n1917 | Февральская революция';
        _answers.text = '1, 2';
      case 'match_pairs':
        _payload.clear();
        _options.text = 'Сущность | Описание';
        _answers.clear();
      case 'image_question':
        _payload.text = 'https://example.com/image.jpg\nОписание изображения';
        _options.text = _choiceOptions();
        _answers.text = '1';
      case 'match_image':
      case 'match_photos':
        _payload.clear();
        _options.text =
            'https://example.com/image.jpg | Описание изображения | Подпись';
        _answers.clear();
      case 'quote_question':
        _payload.text = 'Текст цитаты\nИсточник';
        _options.text = _choiceOptions();
        _answers.text = '1';
      case 'true_false':
        _payload.clear();
        _options.clear();
        _answers.text = 'верно';
      case 'fill_in_blank':
        _payload.text = 'В ____ году ...';
        _options.clear();
        _answers.text = '1905';
      case 'map_point':
        _resetMapEditor();
        _options.clear();
      case 'map_area':
        _resetMapEditor();
        _options.clear();
      case 'multiple_choice':
        _payload.clear();
        _options.text = _choiceOptions();
        _answers.text = '1, 3';
      default:
        _payload.clear();
        _options.text = _choiceOptions();
        _answers.text = '1';
    }
  }

  Future<void> _pickAndUploadImage(String type, {int? photoIndex}) async {
    final client = SessionScope.of(context).client;
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: const ['png', 'jpg', 'jpeg', 'webp'],
      withData: true,
    );
    final file = result?.files.single;
    final bytes = file?.bytes;
    if (file == null || bytes == null) return;
    setState(() {
      _uploadingImage = true;
      if (type == 'match_photos') {
        _uploadingMatchPhotoIndex = photoIndex;
      }
      _error = null;
    });
    try {
      final upload = await MediaApi(client).uploadImage(
        filename: file.name,
        bytes: bytes,
        contentType: _contentTypeForName(file.name),
      );
      if (!mounted) return;
      setState(() {
        if (type == 'image_question') {
          final lines = _nonEmptyLines(_payload.text);
          final alt = lines.length > 1 ? lines.sublist(1).join(' ') : '';
          _payload.text = [upload.url, if (alt.isNotEmpty) alt].join('\n');
        } else if (type == 'match_photos') {
          final lines = _nonEmptyLines(_options.text);
          final rows = lines.map(_MatchPhotoLine.fromLine).toList();
          if (photoIndex != null &&
              photoIndex >= 0 &&
              photoIndex < rows.length) {
            rows[photoIndex] = rows[photoIndex].copyWith(imageUrl: upload.url);
          } else {
            rows.add(_MatchPhotoLine(
                imageUrl: upload.url, alt: 'Изображение', label: 'Подпись'));
          }
          _options.text = rows.map((row) => row.toLine()).join('\n');
        } else if (type == 'match_image') {
          final lines = _nonEmptyLines(_options.text);
          lines.add('${upload.url} | Изображение | Подпись');
          _options.text = lines.join('\n');
        }
        _uploadingImage = false;
        _uploadingMatchPhotoIndex = null;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _uploadingImage = false;
        _uploadingMatchPhotoIndex = null;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _error = 'Не удалось загрузить изображение';
        _uploadingImage = false;
        _uploadingMatchPhotoIndex = null;
      });
    }
  }

  String _contentTypeForName(String name) {
    final lower = name.toLowerCase();
    if (lower.endsWith('.png')) return 'image/png';
    if (lower.endsWith('.webp')) return 'image/webp';
    return 'image/jpeg';
  }

  String _choiceOptions() {
    return 'Вариант A\nВариант B\nВариант C';
  }

  void _resetMapEditor() {
    _mapCenterLat.text = _formatCoord(_defaultMapCenter.latitude);
    _mapCenterLng.text = _formatCoord(_defaultMapCenter.longitude);
    _mapZoom.text = '5';
    _mapTileUrl.text = _defaultTileUrl;
    _mapAttribution.text = _defaultMapAttribution;
    _mapPointRadius.text = '25000';
    _mapAreaCenterRadius.text = '60000';
    _mapAreaTolerance.text = '0.7';
    _mapPointAnswer = null;
    _mapAreaPoints.clear();
    _mapAreaCenterAnswer = null;
    _mapAreaM2 = null;
    _syncMapRawJson();
  }

  void _resetMapEditorFromButton() {
    setState(_resetMapEditor);
  }

  void _copyMapCenterToAnswer() {
    final center = _mapCenter();
    setState(() {
      if (_type.text == 'map_point') {
        _mapPointAnswer = center;
      } else if (_type.text == 'map_area') {
        _mapAreaCenterAnswer = center;
      }
      _syncMapRawJson();
    });
  }

  void _hydrateMapEditor(ChallengeDto challenge) {
    final payload =
        challenge.payload is Map ? challenge.payload as Map : const {};
    final answers =
        challenge.answers is Map ? challenge.answers as Map : const {};
    final center = _latLngFromMap(payload['center']) ?? _defaultMapCenter;
    _mapCenterLat.text = _formatCoord(center.latitude);
    _mapCenterLng.text = _formatCoord(center.longitude);
    _mapZoom.text = _formatPlain(_doubleValue(payload['zoom']) ?? 5);
    _mapTileUrl.text =
        '${payload['tile_url_template'] ?? payload['tileUrlTemplate'] ?? _defaultTileUrl}';
    _mapAttribution.text =
        '${payload['attribution'] ?? _defaultMapAttribution}';
    _mapPointRadius.text =
        _formatPlain(_doubleValue(answers['radius_m']) ?? 25000);
    _mapAreaCenterRadius.text =
        _formatPlain(_doubleValue(answers['center_radius_m']) ?? 60000);
    _mapAreaTolerance.text =
        _formatPlain(_doubleValue(answers['area_tolerance']) ?? 0.7);
    _mapPointAnswer =
        challenge.type == 'map_point' ? _latLngFromMap(answers) : null;
    _mapAreaPoints.clear();
    _mapAreaCenterAnswer =
        challenge.type == 'map_area' ? _latLngFromMap(answers['center']) : null;
    _mapAreaM2 =
        challenge.type == 'map_area' ? _doubleValue(answers['area_m2']) : null;
    _syncMapRawJson();
  }

  void _onMapEditorChanged() {
    setState(_syncMapRawJson);
  }

  void _setMapPointAnswer(LatLng point) {
    setState(() {
      _mapPointAnswer = point;
      _syncMapRawJson();
    });
  }

  void _addMapAreaPoint(LatLng point) {
    setState(() {
      if (_mapAreaPoints.isNotEmpty) {
        final distance =
            const Distance().as(LengthUnit.Meter, _mapAreaPoints.last, point);
        if (distance.abs() < 250) return;
      }
      _mapAreaPoints.add(point);
      _mapAreaCenterAnswer = null;
      _mapAreaM2 = null;
      _syncMapRawJson();
    });
  }

  void _simplifyMapArea() {
    if (_mapAreaPoints.length <= 30) return;
    final step = (_mapAreaPoints.length / 30).ceil();
    final simplified = <LatLng>[];
    for (var i = 0; i < _mapAreaPoints.length; i += step) {
      simplified.add(_mapAreaPoints[i]);
    }
    if (simplified.last != _mapAreaPoints.last) {
      simplified.add(_mapAreaPoints.last);
    }
    final stats = simplified.length >= 3 ? _polygonStats(simplified) : null;
    setState(() {
      _mapAreaPoints
        ..clear()
        ..addAll(simplified);
      _mapAreaCenterAnswer = stats?.center;
      _mapAreaM2 = stats?.areaM2;
      _syncMapRawJson();
    });
  }

  void _clearMapArea() {
    setState(() {
      _mapAreaPoints.clear();
      _mapAreaCenterAnswer = null;
      _mapAreaM2 = null;
      _syncMapRawJson();
    });
  }

  void _finalizeMapArea() {
    if (_mapAreaPoints.length < 3) return;
    final stats = _polygonStats(_mapAreaPoints);
    setState(() {
      _mapAreaCenterAnswer = stats.center;
      _mapAreaM2 = stats.areaM2;
      _syncMapRawJson();
    });
  }

  void _centerMapFromAnswer() {
    final center =
        _type.text == 'map_point' ? _mapPointAnswer : _mapAreaCenterAnswer;
    if (center == null) return;
    setState(() {
      _mapCenterLat.text = _formatCoord(center.latitude);
      _mapCenterLng.text = _formatCoord(center.longitude);
      _syncMapRawJson();
    });
  }

  void _setMatchPhotoLines(List<String> lines) {
    setState(() {
      _options.text = lines.join('\n');
    });
  }

  void _validateMatchPhotosEditor() {
    if (_type.text != 'match_photos') return;
    final rows = _nonEmptyLines(_options.text)
        .map(_MatchPhotoLine.fromLine)
        .where((row) => !row.isBlank)
        .toList();
    if (rows.length < 2) {
      throw const FormatException('match_photos needs at least two pairs');
    }
    for (final row in rows) {
      if (row.label.trim().isEmpty || row.alt.trim().isEmpty) {
        throw const FormatException('match_photos row is incomplete');
      }
      if (_status.text == 'published' && !_isRealMediaUrl(row.imageUrl)) {
        throw const FormatException(
            'published match_photos requires storage image_url');
      }
    }
  }

  void _syncMapRawJson() {
    if (!_isMapType(_type.text)) return;
    _payload.text = _prettyJson(_mapPayloadValue());
    final answer = _mapAnswersOrNull();
    _answers.text = answer == null ? '' : _prettyJson(answer);
  }

  Map<String, dynamic> _mapPayloadValue() {
    return {
      'center': {'lat': _mapCenter().latitude, 'lng': _mapCenter().longitude},
      'zoom': _mapZoomValue(),
      'tile_url_template': _mapTileUrl.text.trim().isEmpty
          ? _defaultTileUrl
          : _mapTileUrl.text.trim(),
      'attribution': _mapAttribution.text.trim().isEmpty
          ? _defaultMapAttribution
          : _mapAttribution.text.trim(),
    };
  }

  dynamic _mapAnswersOrNull() {
    if (_type.text == 'map_point') {
      final point = _mapPointAnswer;
      if (point == null) return null;
      return {
        'lat': point.latitude,
        'lng': point.longitude,
        'radius_m': _mapPointRadiusValue()
      };
    }
    if (_type.text == 'map_area') {
      final center = _mapAreaCenterAnswer;
      final area = _mapAreaM2;
      if (center == null || area == null || area <= 0) return null;
      return {
        'center': {'lat': center.latitude, 'lng': center.longitude},
        'area_m2': area,
        'center_radius_m': _mapAreaCenterRadiusValue(),
        'area_tolerance': _mapAreaToleranceValue(),
      };
    }
    return null;
  }

  dynamic _mapAnswersValue() {
    final answer = _mapAnswersOrNull();
    if (answer == null) {
      throw const FormatException('Map answer is incomplete');
    }
    return answer;
  }

  void _validateMapEditor() {
    if (!_isMapType(_type.text)) return;
    final centerLat = _doubleValue(_mapCenterLat.text);
    final centerLng = _doubleValue(_mapCenterLng.text);
    final zoom = _doubleValue(_mapZoom.text);
    if (centerLat == null ||
        centerLat < -90 ||
        centerLat > 90 ||
        centerLng == null ||
        centerLng < -180 ||
        centerLng > 180) {
      throw const FormatException('Map center is invalid');
    }
    if (zoom == null || zoom < 1 || zoom > 18) {
      throw const FormatException('Map zoom must be between 1 and 18');
    }
    if (_mapTileUrl.text.trim().isEmpty ||
        _mapAttribution.text.trim().isEmpty) {
      throw const FormatException('Map tile settings are incomplete');
    }
    if (_type.text == 'map_point') {
      if (_mapPointAnswer == null) {
        throw const FormatException('Map point is required');
      }
      final radius = _doubleValue(_mapPointRadius.text);
      if (radius == null || radius <= 0) {
        throw const FormatException('Map point radius is required');
      }
    }
    if (_type.text == 'map_area') {
      if (_mapAreaPoints.isNotEmpty && _mapAreaPoints.length < 3) {
        throw const FormatException('Map area needs at least three points');
      }
      if (_mapAreaCenterAnswer == null || _mapAreaM2 == null) {
        throw const FormatException('Map area is required');
      }
      final centerRadius = _doubleValue(_mapAreaCenterRadius.text);
      final areaTolerance = _doubleValue(_mapAreaTolerance.text);
      if (centerRadius == null ||
          centerRadius <= 0 ||
          areaTolerance == null ||
          areaTolerance < 0) {
        throw const FormatException('Map area tolerance is invalid');
      }
    }
  }

  LatLng _mapCenter() {
    return LatLng(
      _doubleValue(_mapCenterLat.text) ?? _defaultMapCenter.latitude,
      _doubleValue(_mapCenterLng.text) ?? _defaultMapCenter.longitude,
    );
  }

  double _mapZoomValue() =>
      (_doubleValue(_mapZoom.text) ?? 5).clamp(1, 18).toDouble();

  double _mapPointRadiusValue() => _doubleValue(_mapPointRadius.text) ?? 25000;

  double _mapAreaCenterRadiusValue() =>
      _doubleValue(_mapAreaCenterRadius.text) ?? 60000;

  double _mapAreaToleranceValue() =>
      _doubleValue(_mapAreaTolerance.text) ?? 0.7;

  LatLng? _latLngFromMap(dynamic raw) {
    if (raw is! Map) return null;
    final lat = _doubleValue(raw['lat']);
    final lng = _doubleValue(raw['lng']);
    if (lat == null || lng == null) return null;
    return LatLng(lat, lng);
  }

  double? _doubleValue(dynamic raw) {
    if (raw is num) return raw.toDouble();
    if (raw is String) return double.tryParse(raw.replaceAll(',', '.'));
    return null;
  }

  _MapAreaStats _polygonStats(List<LatLng> points) {
    final origin = _averageLatLng(points);
    final coords = points
        .map((point) => _ProjectedPoint(
              x: _degToRad(point.longitude - origin.longitude) *
                  6371000 *
                  math.cos(_degToRad(origin.latitude)),
              y: _degToRad(point.latitude - origin.latitude) * 6371000,
            ))
        .toList();
    var twiceArea = 0.0;
    var centroidX = 0.0;
    var centroidY = 0.0;
    for (var i = 0; i < coords.length; i++) {
      final j = (i + 1) % coords.length;
      final cross = coords[i].x * coords[j].y - coords[j].x * coords[i].y;
      twiceArea += cross;
      centroidX += (coords[i].x + coords[j].x) * cross;
      centroidY += (coords[i].y + coords[j].y) * cross;
    }
    if (twiceArea.abs() < 1e-9) {
      return _MapAreaStats(center: origin, areaM2: 0);
    }
    centroidX /= 3 * twiceArea;
    centroidY /= 3 * twiceArea;
    return _MapAreaStats(
      center: LatLng(
        origin.latitude + _radToDeg(centroidY / 6371000),
        origin.longitude +
            _radToDeg(
                centroidX / (6371000 * math.cos(_degToRad(origin.latitude)))),
      ),
      areaM2: twiceArea.abs() / 2,
    );
  }

  LatLng _averageLatLng(List<LatLng> points) {
    var lat = 0.0;
    var lng = 0.0;
    for (final point in points) {
      lat += point.latitude;
      lng += point.longitude;
    }
    return LatLng(lat / points.length, lng / points.length);
  }

  double _degToRad(double value) => value * math.pi / 180;

  double _radToDeg(double value) => value * 180 / math.pi;

  String _formatCoord(double value) => value.toStringAsFixed(6);

  String _formatPlain(double value) {
    if (value == value.roundToDouble()) return value.toStringAsFixed(0);
    return value
        .toStringAsFixed(6)
        .replaceFirst(RegExp(r'0+$'), '')
        .replaceFirst(RegExp(r'\.$'), '');
  }

  void _hydrateEditorFields(ChallengeDto challenge) {
    switch (challenge.type) {
      case 'theory':
        final facts = (challenge.payload['facts'] as List?)
                ?.map((item) => '$item')
                .toList() ??
            [];
        final summary = '${challenge.payload['summary'] ?? ''}'.trim();
        _payload.text = [
          ...facts,
          if (summary.isNotEmpty) '',
          if (summary.isNotEmpty) summary
        ].join('\n');
      case 'timeline':
        final events =
            challenge.options is List ? challenge.options as List : const [];
        _options.text = events.map((item) {
          final map = item is Map ? item : const {};
          return '${map['date'] ?? ''} | ${map['text'] ?? ''}'.trim();
        }).join('\n');
        _answers.text = _answerIndexes(events, challenge.answers).join(', ');
      case 'match_pairs':
        _options.text =
            _pairLines(challenge.options, challenge.answers, imageMode: false)
                .join('\n');
      case 'match_image':
      case 'match_photos':
        _options.text =
            _pairLines(challenge.options, challenge.answers, imageMode: true)
                .join('\n');
      case 'image_question':
        _payload.text =
            '${challenge.payload['image_url'] ?? ''}\n${challenge.payload['alt'] ?? ''}'
                .trim();
        _hydrateChoiceFields(challenge);
      case 'quote_question':
        _payload.text =
            '${challenge.payload['quote'] ?? ''}\n${challenge.payload['source'] ?? ''}'
                .trim();
        _hydrateChoiceFields(challenge);
      case 'true_false':
        _answers.text =
            challenge.answers.contains('false') ? 'неверно' : 'верно';
      case 'fill_in_blank':
        _payload.text = '${challenge.payload['text'] ?? ''}';
        _answers.text = challenge.answers.join('\n');
      case 'map_point':
      case 'map_area':
        _hydrateMapEditor(challenge);
      default:
        _hydrateChoiceFields(challenge);
    }
  }

  void _hydrateChoiceFields(ChallengeDto challenge) {
    final options =
        challenge.options is List ? challenge.options as List : const [];
    _options.text = options
        .map((item) => item is Map ? '${item['text'] ?? ''}' : '$item')
        .join('\n');
    _answers.text = _answerIndexes(options, challenge.answers).join(', ');
  }

  List<String> _pairLines(dynamic options, List<dynamic> answers,
      {required bool imageMode}) {
    if (options is! Map) return const [];
    final photoMode = options['photos'] is List || options['labels'] is List;
    final left = photoMode
        ? (options['photos'] is List ? options['photos'] as List : const [])
        : (options['left'] is List ? options['left'] as List : const []);
    final right = photoMode
        ? (options['labels'] is List ? options['labels'] as List : const [])
        : (options['right'] is List ? options['right'] as List : const []);
    final rightById = {
      for (final item in right)
        if (item is Map) '${item['id']}': item,
    };
    if (answers.isEmpty) {
      return List.generate(
          left.length < right.length ? left.length : right.length, (index) {
        final leftMap = left[index] is Map ? left[index] as Map : const {};
        final rightMap = right[index] is Map ? right[index] as Map : const {};
        return imageMode
            ? '${leftMap['image_url'] ?? ''} | ${leftMap['alt'] ?? ''} | ${rightMap['text'] ?? ''}'
            : '${leftMap['text'] ?? ''} | ${rightMap['text'] ?? ''}';
      });
    }
    return answers.map((answer) {
      final answerMap = answer is Map ? answer : const {};
      final leftAnswerKey = photoMode ? 'photo_id' : 'left_id';
      final rightAnswerKey = photoMode ? 'label_id' : 'right_id';
      final leftMap = left.cast<dynamic>().firstWhere(
            (item) =>
                item is Map && '${item['id']}' == '${answerMap[leftAnswerKey]}',
            orElse: () => const {},
          ) as Map;
      final rightMap = rightById['${answerMap[rightAnswerKey]}'] ?? const {};
      return imageMode
          ? '${leftMap['image_url'] ?? ''} | ${leftMap['alt'] ?? ''} | ${rightMap['text'] ?? ''}'
          : '${leftMap['text'] ?? ''} | ${rightMap['text'] ?? ''}';
    }).toList();
  }

  List<int> _answerIndexes(List<dynamic> options, List<dynamic> answers) {
    final ids = [
      for (final item in options)
        if (item is Map) '${item['id']}',
    ];
    final indexes = <int>[];
    for (final answer in answers) {
      final index = ids.indexOf('$answer');
      if (index >= 0) indexes.add(index + 1);
    }
    return indexes;
  }

  Map<String, dynamic> _payloadValue() {
    final type = _type.text;
    final lines = _nonEmptyLines(_payload.text);
    if (type == 'theory') {
      return {
        'facts':
            lines.length > 1 ? lines.take(lines.length - 1).toList() : lines,
        'summary': lines.length > 1 ? lines.last : '',
      };
    }
    if (type == 'image_question') {
      return {
        'image_url': lines.isNotEmpty ? lines.first : '',
        'alt': lines.length > 1 ? lines.sublist(1).join(' ') : ''
      };
    }
    if (type == 'quote_question') {
      return {
        'quote': lines.isNotEmpty ? lines.first : '',
        'source': lines.length > 1 ? lines.sublist(1).join(' ') : ''
      };
    }
    if (type == 'fill_in_blank') {
      return {'text': _payload.text.trim(), 'placeholder': '____'};
    }
    if (type == 'map_point' || type == 'map_area') {
      return _mapPayloadValue();
    }
    return {};
  }

  dynamic _optionsValue() {
    final type = _type.text;
    if (type == 'theory' ||
        type == 'fill_in_blank' ||
        type == 'map_point' ||
        type == 'map_area') {
      return [];
    }
    if (type == 'true_false') {
      return [
        {'id': 'true', 'text': 'Верно'},
        {'id': 'false', 'text': 'Неверно'},
      ];
    }
    if (type == 'timeline') {
      return _nonEmptyLines(_options.text).asMap().entries.map((entry) {
        final parts = _splitParts(entry.value);
        return {
          'id': _idFor(entry.key),
          'date': parts.isNotEmpty ? parts.first : '',
          'text': parts.length > 1 ? parts.sublist(1).join(' | ') : entry.value
        };
      }).toList();
    }
    if (type == 'match_pairs' ||
        type == 'match_image' ||
        type == 'match_photos') {
      final lines = _nonEmptyLines(_options.text);
      final leftKey = type == 'match_photos' ? 'photos' : 'left';
      final rightKey = type == 'match_photos' ? 'labels' : 'right';
      return {
        leftKey: lines.asMap().entries.map((entry) {
          final parts = _splitParts(entry.value);
          return type == 'match_image' || type == 'match_photos'
              ? {
                  'id': type == 'match_photos'
                      ? 'p${entry.key + 1}'
                      : 'l${entry.key + 1}',
                  'image_url': parts.isNotEmpty ? parts[0] : '',
                  'alt': parts.length > 1 ? parts[1] : ''
                }
              : {
                  'id': 'l${entry.key + 1}',
                  'text': parts.isNotEmpty ? parts[0] : ''
                };
        }).toList(),
        rightKey: lines.asMap().entries.map((entry) {
          final parts = _splitParts(entry.value);
          return {
            'id': type == 'match_photos'
                ? 'l${entry.key + 1}'
                : 'r${entry.key + 1}',
            'text': parts.length > (type == 'match_pairs' ? 1 : 2)
                ? parts[type == 'match_pairs' ? 1 : 2]
                : ''
          };
        }).toList(),
      };
    }
    return _choiceList();
  }

  dynamic _answersValue() {
    final type = _type.text;
    if (type == 'theory') return [];
    if (type == 'match_pairs' || type == 'match_image') {
      final count = _nonEmptyLines(_options.text).length;
      return List.generate(count,
          (index) => {'left_id': 'l${index + 1}', 'right_id': 'r${index + 1}'});
    }
    if (type == 'match_photos') {
      final count = _nonEmptyLines(_options.text).length;
      return List.generate(
          count,
          (index) =>
              {'photo_id': 'p${index + 1}', 'label_id': 'l${index + 1}'});
    }
    if (type == 'true_false') {
      final text = _answers.text.trim().toLowerCase();
      return [text.contains('не') || text == 'false' ? 'false' : 'true'];
    }
    if (type == 'fill_in_blank') return _nonEmptyLines(_answers.text);
    if (type == 'map_point' || type == 'map_area') {
      return _mapAnswersValue();
    }
    final optionIds = _choiceList().map((item) => '${item['id']}').toList();
    final indexes = _parseAnswerIndexes(_answers.text, optionIds.length);
    return [for (final index in indexes) optionIds[index]];
  }

  List<Map<String, String>> _choiceList() {
    return _nonEmptyLines(_options.text)
        .asMap()
        .entries
        .map((entry) => {'id': _idFor(entry.key), 'text': entry.value})
        .toList();
  }

  List<String> _nonEmptyLines(String text) => text
      .split('\n')
      .map((line) => line.trim())
      .where((line) => line.isNotEmpty)
      .toList();

  List<String> _splitParts(String line) =>
      line.split('|').map((part) => part.trim()).toList();

  String _idFor(int index) => String.fromCharCode('a'.codeUnitAt(0) + index);

  String _prettyJson(dynamic value) =>
      const JsonEncoder.withIndent('  ').convert(value);

  List<int> _parseAnswerIndexes(String text, int optionCount) {
    final parts = text
        .split(RegExp(r'[,;\s]+'))
        .map((part) => part.trim().toLowerCase())
        .where((part) => part.isNotEmpty);
    final indexes = <int>[];
    for (final part in parts) {
      final number = int.tryParse(part);
      final index =
          number != null ? number - 1 : part.codeUnitAt(0) - 'a'.codeUnitAt(0);
      if (index >= 0 && index < optionCount && !indexes.contains(index)) {
        indexes.add(index);
      }
    }
    return indexes.isEmpty && optionCount > 0 ? [0] : indexes;
  }

  bool _usesBody(String type) => type == 'theory';
  bool _isMapType(String type) => type == 'map_point' || type == 'map_area';
  bool _usesImageUpload(String type) =>
      type == 'image_question' ||
      type == 'match_image' ||
      type == 'match_photos';
  bool _usesPayload(String type) => [
        'theory',
        'image_question',
        'quote_question',
        'fill_in_blank',
        'map_point',
        'map_area'
      ].contains(type);
  bool _usesOptions(String type) => ![
        'fill_in_blank',
        'theory',
        'true_false',
        'map_point',
        'map_area'
      ].contains(type);
  bool _usesAnswers(String type) =>
      !['theory', 'match_pairs', 'match_image', 'match_photos'].contains(type);
  String _promptLabel(String type) => switch (type) {
        'theory' => 'Заголовок теории',
        _ => 'Вопрос для ученика',
      };
  String _payloadLabel(String type) => switch (type) {
        'theory' => 'Факты и краткое резюме',
        'image_question' => 'Ссылка на изображение и описание',
        'quote_question' => 'Цитата и источник',
        'fill_in_blank' => 'Текст с пропуском',
        'map_point' || 'map_area' => 'JSON карты',
        _ => 'Дополнительные данные',
      };
  String _optionsLabel(String type) => switch (type) {
        'match_pairs' ||
        'match_image' ||
        'match_photos' =>
          'Пары для сопоставления',
        'timeline' => 'События для сортировки',
        'true_false' => 'Варианты “верно / неверно”',
        _ => 'Варианты ответа',
      };
  String _answersLabel(String type) => switch (type) {
        'timeline' => 'Правильный порядок',
        'match_pairs' || 'match_image' || 'match_photos' => 'Правильные пары',
        'true_false' => 'Правильный выбор',
        'fill_in_blank' => 'Допустимые ответы',
        'map_point' => 'JSON правильной точки',
        'map_area' => 'JSON эталонной области',
        _ => 'Правильный ответ',
      };
  String _explanationLabel(String type) => switch (type) {
        'theory' => 'Что важно запомнить',
        _ => 'Объяснение при ошибке',
      };
}

class _MapAuthoringEditor extends StatelessWidget {
  final String type;
  final LatLng center;
  final double zoom;
  final String tileUrlTemplate;
  final String attribution;
  final LatLng? pointAnswer;
  final List<LatLng> areaPoints;
  final LatLng? areaCenter;
  final double? areaM2;
  final bool advancedOpen;
  final TextEditingController centerLatController;
  final TextEditingController centerLngController;
  final TextEditingController zoomController;
  final TextEditingController tileUrlController;
  final TextEditingController attributionController;
  final TextEditingController pointRadiusController;
  final TextEditingController areaCenterRadiusController;
  final TextEditingController areaToleranceController;
  final VoidCallback onChanged;
  final ValueChanged<LatLng> onPointSelected;
  final ValueChanged<LatLng> onAreaPointAdded;
  final VoidCallback onAreaClear;
  final VoidCallback onAreaDone;
  final ValueChanged<bool> onAdvancedChanged;
  final VoidCallback onCenterFromAnswer;
  final VoidCallback onCopyCenterToAnswer;
  final VoidCallback onAreaSimplify;
  final VoidCallback onResetDefaults;

  const _MapAuthoringEditor({
    required this.type,
    required this.center,
    required this.zoom,
    required this.tileUrlTemplate,
    required this.attribution,
    required this.pointAnswer,
    required this.areaPoints,
    required this.areaCenter,
    required this.areaM2,
    required this.advancedOpen,
    required this.centerLatController,
    required this.centerLngController,
    required this.zoomController,
    required this.tileUrlController,
    required this.attributionController,
    required this.pointRadiusController,
    required this.areaCenterRadiusController,
    required this.areaToleranceController,
    required this.onChanged,
    required this.onPointSelected,
    required this.onAreaPointAdded,
    required this.onAreaClear,
    required this.onAreaDone,
    required this.onAdvancedChanged,
    required this.onCenterFromAnswer,
    required this.onCopyCenterToAnswer,
    required this.onAreaSimplify,
    required this.onResetDefaults,
  });

  bool get _isArea => type == 'map_area';

  @override
  Widget build(BuildContext context) {
    final payloadPreview = _prettyJson({
      'center': {'lat': center.latitude, 'lng': center.longitude},
      'zoom': zoom,
      'tile_url_template': tileUrlTemplate,
      'attribution': attribution,
    });
    final answerPreview = _answerPreview();

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppTheme.primary.withValues(alpha: 0.32),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            _isArea ? 'Редактор области на карте' : 'Редактор точки на карте',
            style: GoogleFonts.lato(
                color: AppTheme.textPrimary,
                fontSize: 15,
                fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 10),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: [
              _MapNumberField(
                  controller: centerLatController,
                  label: 'Центр карты lat',
                  width: 126,
                  onChanged: onChanged),
              _MapNumberField(
                  controller: centerLngController,
                  label: 'Центр карты lng',
                  width: 126,
                  onChanged: onChanged),
              _MapNumberField(
                  controller: zoomController,
                  label: 'Zoom',
                  width: 90,
                  onChanged: onChanged),
              if (!_isArea)
                _MapNumberField(
                    controller: pointRadiusController,
                    label: 'Радиус, м',
                    width: 122,
                    onChanged: onChanged),
              if (_isArea)
                _MapNumberField(
                    controller: areaCenterRadiusController,
                    label: 'Допуск центра, м',
                    width: 122,
                    onChanged: onChanged),
              if (_isArea)
                _MapNumberField(
                    controller: areaToleranceController,
                    label: 'Допуск площади',
                    width: 112,
                    onChanged: onChanged),
            ],
          ),
          const SizedBox(height: 12),
          SizedBox(
            height: 340,
            child: ClipRRect(
              borderRadius: BorderRadius.circular(14),
              child: FlutterMap(
                key: ValueKey(
                    '${type}_${center.latitude}_${center.longitude}_$zoom'),
                options: MapOptions(
                  initialCenter: center,
                  initialZoom: zoom,
                  interactionOptions: InteractionOptions(
                      flags:
                          _isArea ? InteractiveFlag.none : InteractiveFlag.all),
                  onTap: !_isArea ? (_, point) => onPointSelected(point) : null,
                  onPointerDown:
                      _isArea ? (_, point) => onAreaPointAdded(point) : null,
                  onPointerMove:
                      _isArea ? (_, point) => onAreaPointAdded(point) : null,
                ),
                children: [
                  TileLayer(
                    urlTemplate: tileUrlTemplate,
                    userAgentPackageName: 'history_app',
                  ),
                  if (_isArea && areaPoints.length >= 3)
                    PolygonLayer(
                      polygons: [
                        Polygon(
                          points: areaPoints,
                          color: AppTheme.accent.withValues(alpha: 0.24),
                          borderColor: AppTheme.accent,
                          borderStrokeWidth: 3,
                        ),
                      ],
                    ),
                  if (_isArea && areaPoints.length >= 2)
                    PolylineLayer(
                      polylines: [
                        Polyline(
                            points: areaPoints,
                            color: AppTheme.accent,
                            strokeWidth: 3),
                      ],
                    ),
                  MarkerLayer(
                    markers: [
                      if (pointAnswer != null)
                        Marker(
                          point: pointAnswer!,
                          width: 46,
                          height: 46,
                          child: const Icon(Icons.location_on,
                              color: AppTheme.accent, size: 42),
                        ),
                      if (_isArea && areaCenter != null)
                        Marker(
                          point: areaCenter!,
                          width: 38,
                          height: 38,
                          child: const Icon(Icons.center_focus_strong,
                              color: AppTheme.correct, size: 32),
                        ),
                    ],
                  ),
                  Align(
                    alignment: Alignment.bottomRight,
                    child: Container(
                      margin: const EdgeInsets.all(6),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 3),
                      decoration: BoxDecoration(
                          color: AppTheme.surface.withValues(alpha: 0.86),
                          borderRadius: BorderRadius.circular(6)),
                      child: Text(attribution,
                          style: GoogleFonts.lato(
                              color: AppTheme.textPrimary, fontSize: 10)),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),
          _mapActions(),
          const SizedBox(height: 8),
          Text(
            _statusText(),
            style:
                GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 12),
          ),
          ExpansionTile(
            initiallyExpanded: advancedOpen,
            onExpansionChanged: onAdvancedChanged,
            tilePadding: EdgeInsets.zero,
            iconColor: AppTheme.accent,
            collapsedIconColor: AppTheme.textSecondary,
            title: Text('Дополнительно и JSON preview',
                style: GoogleFonts.lato(
                    color: AppTheme.textPrimary, fontWeight: FontWeight.bold)),
            children: [
              _TextInput(
                  controller: tileUrlController,
                  label: 'Tile URL template',
                  onChanged: (_) => onChanged()),
              const SizedBox(height: 10),
              _TextInput(
                  controller: attributionController,
                  label: 'Attribution',
                  onChanged: (_) => onChanged()),
              const SizedBox(height: 10),
              _JsonPreview(title: 'payload', value: payloadPreview),
              const SizedBox(height: 10),
              _JsonPreview(
                  title: 'answers',
                  value: answerPreview ?? 'Ответ ещё не выбран'),
            ],
          ),
        ],
      ),
    );
  }

  Widget _mapActions() {
    if (_isArea) {
      return Wrap(
        spacing: 10,
        runSpacing: 8,
        children: [
          OutlinedButton.icon(
            onPressed:
                areaPoints.isEmpty && areaCenter == null ? null : onAreaClear,
            icon: const Icon(Icons.backspace_outlined),
            label: const ButtonLabel('Очистить'),
            style: OutlinedButton.styleFrom(
                side: const BorderSide(color: AppTheme.accent)),
          ),
          ElevatedButton.icon(
            onPressed: areaPoints.length < 3 ? null : onAreaDone,
            icon: const Icon(Icons.done),
            label: const ButtonLabel('Готово'),
            style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.accent,
                foregroundColor: AppTheme.onAccent),
          ),
          OutlinedButton.icon(
            onPressed: areaCenter == null ? null : onCenterFromAnswer,
            icon: const Icon(Icons.center_focus_strong),
            label: const ButtonLabel('Центр из области'),
            style: OutlinedButton.styleFrom(
                side: const BorderSide(color: AppTheme.accent)),
          ),
          OutlinedButton.icon(
            onPressed: onCopyCenterToAnswer,
            icon: const Icon(Icons.my_location_outlined),
            label: const ButtonLabel('Ответ из центра'),
            style: OutlinedButton.styleFrom(
                side: const BorderSide(color: AppTheme.accent)),
          ),
          OutlinedButton.icon(
            onPressed: areaPoints.length <= 30 ? null : onAreaSimplify,
            icon: const Icon(Icons.compress_outlined),
            label: const ButtonLabel('Упростить'),
            style: OutlinedButton.styleFrom(
                side: const BorderSide(color: AppTheme.accent)),
          ),
          OutlinedButton.icon(
            onPressed: onResetDefaults,
            icon: const Icon(Icons.restart_alt_outlined),
            label: const ButtonLabel('OSM default'),
            style: OutlinedButton.styleFrom(
                side: const BorderSide(color: AppTheme.accent)),
          ),
        ],
      );
    }
    return Wrap(
      spacing: 10,
      runSpacing: 8,
      children: [
        OutlinedButton.icon(
          onPressed: pointAnswer == null ? null : onCenterFromAnswer,
          icon: const Icon(Icons.center_focus_strong),
          label: const ButtonLabel('Центр из точки'),
          style: OutlinedButton.styleFrom(
              side: const BorderSide(color: AppTheme.accent)),
        ),
        OutlinedButton.icon(
          onPressed: onCopyCenterToAnswer,
          icon: const Icon(Icons.my_location_outlined),
          label: const ButtonLabel('Ответ из центра'),
          style: OutlinedButton.styleFrom(
              side: const BorderSide(color: AppTheme.accent)),
        ),
        OutlinedButton.icon(
          onPressed: onResetDefaults,
          icon: const Icon(Icons.restart_alt_outlined),
          label: const ButtonLabel('OSM default'),
          style: OutlinedButton.styleFrom(
              side: const BorderSide(color: AppTheme.accent)),
        ),
      ],
    );
  }

  String _statusText() {
    if (_isArea) {
      if (areaCenter == null || areaM2 == null) {
        return areaPoints.length < 3
            ? 'Проведи контур по карте минимум из 3 точек.'
            : 'Нажми “Готово”, чтобы рассчитать центр и площадь.';
      }
      return 'Центр: ${areaCenter!.latitude.toStringAsFixed(5)}, ${areaCenter!.longitude.toStringAsFixed(5)} · площадь: ${areaM2!.round()} м² · точек: ${areaPoints.length}';
    }
    if (pointAnswer == null) {
      return 'Кликни по карте, чтобы поставить правильную точку.';
    }
    return 'Точка: ${pointAnswer!.latitude.toStringAsFixed(5)}, ${pointAnswer!.longitude.toStringAsFixed(5)}';
  }

  String? _answerPreview() {
    if (_isArea) {
      if (areaCenter == null || areaM2 == null) return null;
      return _prettyJson({
        'center': {'lat': areaCenter!.latitude, 'lng': areaCenter!.longitude},
        'area_m2': areaM2,
        'center_radius_m': double.tryParse(
                areaCenterRadiusController.text.replaceAll(',', '.')) ??
            60000,
        'area_tolerance': double.tryParse(
                areaToleranceController.text.replaceAll(',', '.')) ??
            0.7,
      });
    }
    if (pointAnswer == null) return null;
    return _prettyJson({
      'lat': pointAnswer!.latitude,
      'lng': pointAnswer!.longitude,
      'radius_m':
          double.tryParse(pointRadiusController.text.replaceAll(',', '.')) ??
              25000,
    });
  }

  String _prettyJson(dynamic value) =>
      const JsonEncoder.withIndent('  ').convert(value);
}

class _MapNumberField extends StatelessWidget {
  final TextEditingController controller;
  final String label;
  final double width;
  final VoidCallback onChanged;

  const _MapNumberField(
      {required this.controller,
      required this.label,
      required this.width,
      required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: width,
      child: TextField(
        controller: controller,
        keyboardType:
            const TextInputType.numberWithOptions(decimal: true, signed: true),
        onChanged: (_) => onChanged(),
        style: GoogleFonts.lato(color: AppTheme.textPrimary),
        decoration: _inputDecoration(label),
      ),
    );
  }
}

class _JsonPreview extends StatelessWidget {
  final String title;
  final String value;

  const _JsonPreview({required this.title, required this.value});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: AppTheme.primary.withValues(alpha: 0.45),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title,
              style: GoogleFonts.lato(
                  color: AppTheme.accent, fontWeight: FontWeight.bold)),
          const SizedBox(height: 6),
          SelectableText(value,
              style: GoogleFonts.robotoMono(
                  color: AppTheme.textSecondary, fontSize: 12)),
        ],
      ),
    );
  }
}

class _MatchPhotosAuthoringEditor extends StatelessWidget {
  final List<String> lines;
  final bool published;
  final int? uploadingIndex;
  final ValueChanged<int> onUpload;
  final ValueChanged<List<String>> onChanged;

  const _MatchPhotosAuthoringEditor({
    required this.lines,
    required this.published,
    required this.uploadingIndex,
    required this.onUpload,
    required this.onChanged,
  });

  List<_MatchPhotoLine> get _rows {
    final rows = lines.map(_MatchPhotoLine.fromLine).toList();
    return rows.isEmpty ? [_MatchPhotoLine.empty()] : rows;
  }

  @override
  Widget build(BuildContext context) {
    final rows = _rows;
    final hasUnsafeImage =
        rows.any((row) => !_isRealMediaUrl(row.imageUrl) && !row.isBlank);
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppTheme.primary.withValues(alpha: 0.32),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  'Фото-пары',
                  style: GoogleFonts.lato(
                    color: AppTheme.textPrimary,
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              OutlinedButton.icon(
                onPressed: () => _emit([
                  ...rows.where((row) => !row.isBlank),
                  _MatchPhotoLine.empty()
                ]),
                icon: const Icon(Icons.add_photo_alternate_outlined),
                label: const ButtonLabel('Добавить'),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            'Каждая карточка сохраняется как photo_id -> label_id. Для published нужен image_url; до загрузки фото оставляй draft.',
            style: GoogleFonts.lato(
              color: AppTheme.textSecondary,
              fontSize: 12,
            ),
          ),
          if (published && hasUnsafeImage) ...[
            const SizedBox(height: 8),
            Text(
              'Published match_photos без storage image_url будет заблокирован при сохранении.',
              style: GoogleFonts.lato(
                color: AppTheme.error,
                fontWeight: FontWeight.bold,
                fontSize: 12,
              ),
            ),
          ],
          const SizedBox(height: 12),
          ...rows.asMap().entries.map((entry) {
            final index = entry.key;
            final row = entry.value;
            return Padding(
              padding:
                  EdgeInsets.only(bottom: index == rows.length - 1 ? 0 : 12),
              child: _MatchPhotoCard(
                index: index,
                row: row,
                canRemove: rows.length > 1,
                uploading: uploadingIndex == index,
                onChanged: (next) => _replace(rows, index, next),
                onUpload: () => onUpload(index),
                onRemove: () => _remove(rows, index),
              ),
            );
          }),
        ],
      ),
    );
  }

  void _replace(List<_MatchPhotoLine> rows, int index, _MatchPhotoLine row) {
    final next = [...rows];
    next[index] = row;
    _emit(next);
  }

  void _remove(List<_MatchPhotoLine> rows, int index) {
    final next = [...rows]..removeAt(index);
    _emit(next.isEmpty ? [_MatchPhotoLine.empty()] : next);
  }

  void _emit(List<_MatchPhotoLine> rows) {
    onChanged(rows.map((row) => row.toLine()).toList());
  }
}

class _MatchPhotoCard extends StatelessWidget {
  final int index;
  final _MatchPhotoLine row;
  final bool canRemove;
  final bool uploading;
  final ValueChanged<_MatchPhotoLine> onChanged;
  final VoidCallback onUpload;
  final VoidCallback onRemove;

  const _MatchPhotoCard({
    required this.index,
    required this.row,
    required this.canRemove,
    required this.uploading,
    required this.onChanged,
    required this.onUpload,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppTheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    Chip(
                      label: Text('photo_id: p${index + 1}',
                          style: GoogleFonts.lato(color: AppTheme.textPrimary)),
                      backgroundColor: AppTheme.primary.withValues(alpha: 0.4),
                    ),
                    Chip(
                      label: Text('label_id: l${index + 1}',
                          style: GoogleFonts.lato(color: AppTheme.textPrimary)),
                      backgroundColor: AppTheme.primary.withValues(alpha: 0.4),
                    ),
                  ],
                ),
              ),
              IconButton(
                onPressed: canRemove ? onRemove : null,
                icon: const Icon(Icons.delete_outline),
                color: AppTheme.error,
              ),
            ],
          ),
          const SizedBox(height: 10),
          _photoPreview(),
          const SizedBox(height: 10),
          OutlinedButton.icon(
            onPressed: uploading ? null : onUpload,
            icon: uploading
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.upload_file_outlined),
            label: ButtonLabel(
                uploading ? 'Загрузка...' : 'Загрузить / заменить фото'),
          ),
          const SizedBox(height: 10),
          TextFormField(
            initialValue: row.imageUrl,
            onChanged: (value) => onChanged(row.copyWith(imageUrl: value)),
            style: GoogleFonts.lato(color: AppTheme.textPrimary),
            decoration: _inputDecoration('image_url'),
          ),
          const SizedBox(height: 10),
          TextFormField(
            initialValue: row.alt,
            onChanged: (value) => onChanged(row.copyWith(alt: value)),
            style: GoogleFonts.lato(color: AppTheme.textPrimary),
            decoration: _inputDecoration('Alt / что на фото'),
          ),
          const SizedBox(height: 10),
          TextFormField(
            initialValue: row.label,
            onChanged: (value) => onChanged(row.copyWith(label: value)),
            style: GoogleFonts.lato(color: AppTheme.textPrimary),
            decoration: _inputDecoration('Подпись для сопоставления'),
          ),
        ],
      ),
    );
  }

  Widget _photoPreview() {
    final url = row.imageUrl.trim();
    if (url.isEmpty) return _photoFallback(row.alt);
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: Image.network(
        url,
        height: 120,
        width: double.infinity,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) => _photoFallback(row.alt),
      ),
    );
  }

  Widget _photoFallback(String alt) {
    return Container(
      height: 96,
      width: double.infinity,
      alignment: Alignment.center,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppTheme.primary.withValues(alpha: 0.45),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Text(
        alt.trim().isEmpty ? 'Изображение ещё не задано' : alt,
        textAlign: TextAlign.center,
        style: GoogleFonts.lato(color: AppTheme.textSecondary),
      ),
    );
  }
}

class _MatchPhotoLine {
  final String imageUrl;
  final String alt;
  final String label;

  const _MatchPhotoLine({
    required this.imageUrl,
    required this.alt,
    required this.label,
  });

  factory _MatchPhotoLine.empty() =>
      const _MatchPhotoLine(imageUrl: '', alt: '', label: '');

  factory _MatchPhotoLine.fromLine(String line) {
    final parts = line.split('|').map((part) => part.trim()).toList();
    return _MatchPhotoLine(
      imageUrl: parts.isNotEmpty ? parts[0] : '',
      alt: parts.length > 1 ? parts[1] : '',
      label: parts.length > 2 ? parts.sublist(2).join(' | ') : '',
    );
  }

  bool get isBlank =>
      imageUrl.trim().isEmpty && alt.trim().isEmpty && label.trim().isEmpty;

  _MatchPhotoLine copyWith({String? imageUrl, String? alt, String? label}) {
    return _MatchPhotoLine(
      imageUrl: imageUrl ?? this.imageUrl,
      alt: alt ?? this.alt,
      label: label ?? this.label,
    );
  }

  String toLine() => '$imageUrl | $alt | $label';
}

class _MapAreaStats {
  final LatLng center;
  final double areaM2;

  const _MapAreaStats({required this.center, required this.areaM2});
}

class _ProjectedPoint {
  final double x;
  final double y;

  const _ProjectedPoint({required this.x, required this.y});
}

class _FormDialog extends StatelessWidget {
  final String title;
  final List<Widget> fields;
  final VoidCallback onSave;
  final String? error;

  const _FormDialog(
      {required this.title,
      required this.fields,
      required this.onSave,
      this.error});

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: AppTheme.surface,
      insetPadding: const EdgeInsets.all(16),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxHeight: 680),
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  Expanded(
                      child: Text(title,
                          style: GoogleFonts.playfairDisplay(
                              color: AppTheme.textPrimary,
                              fontSize: 20,
                              fontWeight: FontWeight.bold))),
                  IconButton(
                      onPressed: () => Navigator.pop(context),
                      icon: Icon(Icons.close, color: AppTheme.textSecondary)),
                ],
              ),
            ),
            Expanded(
              child: ListView.separated(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemBuilder: (_, i) => fields[i],
                separatorBuilder: (_, __) => const SizedBox(height: 12),
                itemCount: fields.length,
              ),
            ),
            if (error != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Text(error!,
                    style: GoogleFonts.lato(
                        color: AppTheme.error, fontWeight: FontWeight.bold)),
              ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                    onPressed: onSave, child: const ButtonLabel('Сохранить')),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _TextInput extends StatelessWidget {
  final TextEditingController controller;
  final String label;
  final int maxLines;
  final TextInputType? keyboardType;
  final ValueChanged<String>? onChanged;

  const _TextInput(
      {required this.controller,
      required this.label,
      this.maxLines = 1,
      this.keyboardType,
      this.onChanged});

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      maxLines: maxLines,
      keyboardType: keyboardType,
      onChanged: onChanged,
      style: GoogleFonts.lato(color: AppTheme.textPrimary),
      decoration: _inputDecoration(label),
    );
  }
}

InputDecoration _inputDecoration(String label) {
  return InputDecoration(
    labelText: label,
    labelStyle: GoogleFonts.lato(color: AppTheme.textSecondary),
    filled: true,
    fillColor: AppTheme.primary.withValues(alpha: 0.35),
    enabledBorder: OutlineInputBorder(
        borderSide: BorderSide(color: AppTheme.cardBg),
        borderRadius: BorderRadius.circular(8)),
    focusedBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: AppTheme.accent),
        borderRadius: BorderRadius.circular(8)),
  );
}

class _ErrorBanner extends StatelessWidget {
  final String message;

  const _ErrorBanner(this.message);

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
          color: AppTheme.error.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.error.withValues(alpha: 0.4))),
      child: ResponsiveText(message,
          style: GoogleFonts.lato(
              color: AppTheme.error, fontWeight: FontWeight.bold)),
    );
  }
}

class _InfoBox extends StatelessWidget {
  final String title;
  final String subtitle;

  const _InfoBox(this.title, this.subtitle);

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
          color: AppTheme.surface,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.cardBg)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ResponsiveText(title,
              style: GoogleFonts.lato(
                  color: AppTheme.textPrimary, fontWeight: FontWeight.bold)),
          const SizedBox(height: 4),
          ResponsiveText(subtitle,
              style: GoogleFonts.lato(color: AppTheme.textSecondary)),
        ],
      ),
    );
  }
}

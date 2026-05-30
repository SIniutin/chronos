import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../api/api_client.dart';
import '../api/auth_api.dart';
import '../api/content_api.dart';
import '../state/session_controller.dart';
import '../theme/app_theme.dart';

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
    final user = session.currentUser;
    final tabs = [
      _PanelTab('Контент', Icons.account_tree_outlined),
      _PanelTab('Публикация', Icons.verified_outlined),
      if (user?.isAdmin == true) _PanelTab('Пользователи', Icons.admin_panel_settings_outlined),
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
                avatar: Icon(tabs[i].icon, size: 16, color: _tab == i ? AppTheme.onAccent : AppTheme.accent),
                label: Text(tabs[i].label),
                onSelected: (_) => setState(() => _tab = i),
                selectedColor: AppTheme.accent,
                backgroundColor: AppTheme.surface,
                labelStyle: GoogleFonts.lato(color: _tab == i ? AppTheme.onAccent : AppTheme.textPrimary),
              ),
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemCount: tabs.length,
            ),
          ),
          Expanded(
            child: switch (tabs[_tab].label) {
              'Контент' => const _ContentTreeEditor(mode: _ContentMode.edit),
              'Публикация' => const _ContentTreeEditor(mode: _ContentMode.review),
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
      'quote_question' => 'Вопрос по цитате',
      'true_false' => 'Верно / неверно',
      'fill_in_blank' => 'Заполнить пропуск',
      _ => type,
    };

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
  CourseDto? _course;
  SectionDto? _section;
  UnitDto? _unit;
  SkillDto? _skill;
  bool _loading = true;
  String? _error;

  bool get _canEdit => SessionScope.of(context).currentUser?.canEditContent == true && widget.mode == _ContentMode.edit;
  bool get _canPublish => SessionScope.of(context).currentUser?.canReviewContent == true;
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
          _challenges = [];
        });
      });

  Future<void> _selectUnit(UnitDto unit) => _guard(() async {
        final skills = await _api.listSkills(unit.id);
        setState(() {
          _unit = unit;
          _skill = null;
          _skills = skills;
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
      return const Center(child: CircularProgressIndicator(color: AppTheme.accent));
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
                onPublish: _canPublish ? () => _transition('courses', course.id, publish: true) : null,
                onArchive: _canArchive ? () => _transition('courses', course.id, publish: false) : null,
              )),
          if (_course != null) ...[
            _HeaderRow(title: 'Разделы', canAdd: _canEdit, onAdd: () => _editSection(null)),
            ..._sections.map((section) => _EntityTile(
                  title: section.theme,
                  subtitle: section.description,
                  status: section.status,
                  selected: _section?.id == section.id,
                  onTap: () => _selectSection(section),
                  onEdit: _canEdit ? () => _editSection(section) : null,
                  onPublish: _canPublish ? () => _transition('sections', section.id, publish: true) : null,
                  onArchive: _canArchive ? () => _transition('sections', section.id, publish: false) : null,
                )),
          ],
          if (_section != null) ...[
            _HeaderRow(title: 'Уроки', canAdd: _canEdit, onAdd: () => _editUnit(null)),
            ..._units.map((unit) => _EntityTile(
                  title: unit.title,
                  subtitle: 'Порядок показа: ${unit.position}',
                  status: unit.status,
                  selected: _unit?.id == unit.id,
                  onTap: () => _selectUnit(unit),
                  onEdit: _canEdit ? () => _editUnit(unit) : null,
                  onPublish: _canPublish ? () => _transition('units', unit.id, publish: true) : null,
                  onArchive: _canArchive ? () => _transition('units', unit.id, publish: false) : null,
                )),
          ],
          if (_unit != null) ...[
            _HeaderRow(title: 'Навыки', canAdd: _canEdit, onAdd: () => _editSkill(null)),
            ..._skills.map((skill) => _EntityTile(
                  title: '${skill.icon} ${skill.title}',
                  subtitle: 'Порядок показа: ${skill.position}',
                  status: skill.status,
                  selected: _skill?.id == skill.id,
                  onTap: () => _selectSkill(skill),
                  onEdit: _canEdit ? () => _editSkill(skill) : null,
                  onPublish: _canPublish ? () => _transition('skills', skill.id, publish: true) : null,
                  onArchive: _canArchive ? () => _transition('skills', skill.id, publish: false) : null,
                )),
          ],
          if (_skill != null) ...[
            _HeaderRow(title: 'Вопросы и теория', canAdd: _canEdit, onAdd: () => _editChallenge(null)),
            ..._challenges.map((challenge) => _EntityTile(
                  title: challenge.prompt,
                  subtitle: '${_challengeTypeLabel(challenge.type)} · ${_difficultyLabel(challenge.difficulty)}',
                  status: challenge.status,
                  selected: false,
                  onTap: _canEdit ? () => _editChallenge(challenge) : () {},
                  onEdit: _canEdit ? () => _editChallenge(challenge) : null,
                  onPublish: _canPublish ? () => _transition('challenges', challenge.id, publish: true) : null,
                  onArchive: _canArchive ? () => _transition('challenges', challenge.id, publish: false) : null,
                )),
          ],
        ],
      ),
    );
  }

  Future<void> _transition(String entity, String id, {required bool publish}) => _guard(() async {
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

  const _HeaderRow({required this.title, required this.canAdd, required this.onAdd});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 18, bottom: 8),
      child: Row(
        children: [
          Expanded(
            child: Text(title, style: GoogleFonts.playfairDisplay(color: AppTheme.textPrimary, fontSize: 20, fontWeight: FontWeight.bold)),
          ),
          if (canAdd)
            IconButton(
              tooltip: 'Добавить',
              onPressed: onAdd,
              icon: const Icon(Icons.add_circle_outline, color: AppTheme.accent),
            ),
        ],
      ),
    );
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
    return Card(
      color: selected ? AppTheme.cardBg : AppTheme.surface,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8), side: BorderSide(color: selected ? AppTheme.accent : AppTheme.cardBg)),
      child: ListTile(
        onTap: onTap,
        title: Text(title.isEmpty ? 'Без названия' : title, style: GoogleFonts.lato(color: AppTheme.textPrimary, fontWeight: FontWeight.bold)),
        subtitle: Text('$subtitle · ${_statusLabel(status)}', style: GoogleFonts.lato(color: AppTheme.textSecondary)),
        trailing: Wrap(
          spacing: 2,
          children: [
            if (onEdit != null) IconButton(tooltip: 'Редактировать', onPressed: onEdit, icon: const Icon(Icons.edit_outlined, color: AppTheme.accent)),
            if (onPublish != null) IconButton(tooltip: 'Опубликовать', onPressed: onPublish, icon: const Icon(Icons.verified_outlined, color: AppTheme.success)),
            if (onArchive != null) IconButton(tooltip: 'В архив', onPressed: onArchive, icon: const Icon(Icons.archive_outlined, color: AppTheme.wrong)),
          ],
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
    final auth = AuthApi(SessionScope.of(context).client);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text('Смена роли', style: GoogleFonts.playfairDisplay(color: AppTheme.textPrimary, fontSize: 22, fontWeight: FontWeight.bold)),
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
          label: const Text('Найти пользователя'),
        ),
        if (_found != null) ...[
          const SizedBox(height: 18),
          _InfoBox('${_found!.login} · ${_found!.email}', 'Текущая роль: ${_roleLabel(_found!.role)}'),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            value: _role,
            dropdownColor: AppTheme.surface,
            decoration: _inputDecoration('Новая роль'),
            items: const ['student', 'content_editor', 'content_reviewer', 'admin']
                .map((role) => DropdownMenuItem(value: role, child: Text(_roleLabel(role))))
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
            label: const Text('Сохранить роль'),
          ),
        ],
        if (_message != null) ...[
          const SizedBox(height: 14),
          Text(_message!, style: GoogleFonts.lato(color: AppTheme.accent, fontWeight: FontWeight.bold)),
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
      onSave: () => Navigator.pop(context, CourseDto(id: course?.id ?? '', sourceLang: _source.text, targetLang: _target.text, title: _title.text, status: course?.status ?? 'draft')),
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
        _TextInput(controller: _description, label: 'Краткое описание', maxLines: 3),
        _TextInput(controller: _position, label: 'Порядок показа', keyboardType: TextInputType.number),
      ],
      onSave: () => Navigator.pop(context, SectionDto(id: section?.id ?? '', courseId: courseId, theme: _theme.text, description: _description.text, position: int.tryParse(_position.text) ?? 1, status: section?.status ?? 'draft')),
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
        _TextInput(controller: _position, label: 'Порядок показа', keyboardType: TextInputType.number),
      ],
      onSave: () => Navigator.pop(context, UnitDto(id: unit?.id ?? '', sectionId: sectionId, title: _title.text, position: int.tryParse(_position.text) ?? 1, status: unit?.status ?? 'draft')),
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
        _TextInput(controller: _position, label: 'Порядок показа', keyboardType: TextInputType.number),
      ],
      onSave: () => Navigator.pop(context, SkillDto(id: skill?.id ?? '', unitId: unitId, title: _title.text, icon: _icon.text, position: int.tryParse(_position.text) ?? 1, status: skill?.status ?? 'draft')),
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
    'quote_question',
    'true_false',
    'fill_in_blank',
  ];
  static const _difficulties = ['easy', 'medium', 'hard'];

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
  String? _error;

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
    if (challenge == null) {
      _applyTemplate(_type.text);
    } else {
      _hydrateEditorFields(challenge);
    }
  }

  @override
  Widget build(BuildContext context) {
    final type = _type.text;
    return _FormDialog(
      title: widget.challenge == null ? 'Новый вопрос или теория' : 'Редактировать материал',
      error: _error,
      fields: [
        DropdownButtonFormField<String>(
          value: _types.contains(type) ? type : 'single_choice',
          dropdownColor: AppTheme.surface,
          decoration: _inputDecoration('Тип материала'),
          items: _types.map((type) => DropdownMenuItem(value: type, child: Text(_challengeTypeLabel(type)))).toList(),
          onChanged: (value) {
            if (value == null) return;
            setState(() {
              _type.text = value;
              _applyTemplate(value);
            });
          },
        ),
        DropdownButtonFormField<String>(
          value: _difficulties.contains(_difficulty.text) ? _difficulty.text : 'easy',
          dropdownColor: AppTheme.surface,
          decoration: _inputDecoration('Сложность'),
          items: _difficulties.map((item) => DropdownMenuItem(value: item, child: Text(_difficultyLabel(item)))).toList(),
          onChanged: (value) => setState(() => _difficulty.text = value ?? 'easy'),
        ),
        _TextInput(controller: _tags, label: 'Теги через запятую'),
        _TextInput(controller: _level, label: 'Уровень внутри навыка', keyboardType: TextInputType.number),
        _TextInput(controller: _lessonCount, label: 'Количество шагов урока', keyboardType: TextInputType.number),
        _TextInput(controller: _prompt, label: _promptLabel(type), maxLines: 3),
        if (_usesBody(type)) _TextInput(controller: _body, label: 'Текст теории', maxLines: 4),
        if (_usesPayload(type)) _TextInput(controller: _payload, label: _payloadLabel(type), maxLines: 5),
        if (_usesOptions(type)) _TextInput(controller: _options, label: _optionsLabel(type), maxLines: 5),
        if (_usesAnswers(type)) _TextInput(controller: _answers, label: _answersLabel(type), maxLines: 5),
        _TextInput(controller: _explanation, label: _explanationLabel(type), maxLines: 4),
        _TextInput(controller: _position, label: 'Порядок показа', keyboardType: TextInputType.number),
      ],
      onSave: () {
        try {
          Navigator.pop(
            context,
            ChallengeDto(
              id: widget.challenge?.id ?? '',
              skillId: widget.skillId,
              type: _type.text,
              difficulty: _difficulty.text,
              tags: _tags.text.split(',').map((tag) => tag.trim()).where((tag) => tag.isNotEmpty).toList(),
              level: int.tryParse(_level.text) ?? 1,
              lessonCount: int.tryParse(_lessonCount.text) ?? 1,
              prompt: _prompt.text,
              body: _body.text,
              payload: _payloadValue(),
              options: _optionsValue(),
              answers: _answersValue(),
              explanation: _explanation.text,
              position: int.tryParse(_position.text) ?? 1,
              status: widget.challenge?.status ?? 'draft',
            ),
          );
        } catch (_) {
          setState(() => _error = 'Проверьте поля с вариантами, ответами и дополнительными данными');
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
        _options.text = '1905 | Первая русская революция\n1917 | Февральская революция';
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
        _payload.clear();
        _options.text = 'https://example.com/image.jpg | Описание изображения | Подпись';
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

  String _choiceOptions() {
    return 'Вариант A\nВариант B\nВариант C';
  }

  void _hydrateEditorFields(ChallengeDto challenge) {
    switch (challenge.type) {
      case 'theory':
        final facts = (challenge.payload['facts'] as List?)?.map((item) => '$item').toList() ?? [];
        final summary = '${challenge.payload['summary'] ?? ''}'.trim();
        _payload.text = [...facts, if (summary.isNotEmpty) '', if (summary.isNotEmpty) summary].join('\n');
      case 'timeline':
        final events = challenge.options is List ? challenge.options as List : const [];
        _options.text = events.map((item) {
          final map = item is Map ? item : const {};
          return '${map['date'] ?? ''} | ${map['text'] ?? ''}'.trim();
        }).join('\n');
        _answers.text = _answerIndexes(events, challenge.answers).join(', ');
      case 'match_pairs':
        _options.text = _pairLines(challenge.options, challenge.answers, imageMode: false).join('\n');
      case 'match_image':
        _options.text = _pairLines(challenge.options, challenge.answers, imageMode: true).join('\n');
      case 'image_question':
        _payload.text = '${challenge.payload['image_url'] ?? ''}\n${challenge.payload['alt'] ?? ''}'.trim();
        _hydrateChoiceFields(challenge);
      case 'quote_question':
        _payload.text = '${challenge.payload['quote'] ?? ''}\n${challenge.payload['source'] ?? ''}'.trim();
        _hydrateChoiceFields(challenge);
      case 'true_false':
        _answers.text = challenge.answers.contains('false') ? 'неверно' : 'верно';
      case 'fill_in_blank':
        _payload.text = '${challenge.payload['text'] ?? ''}';
        _answers.text = challenge.answers.join('\n');
      default:
        _hydrateChoiceFields(challenge);
    }
  }

  void _hydrateChoiceFields(ChallengeDto challenge) {
    final options = challenge.options is List ? challenge.options as List : const [];
    _options.text = options.map((item) => item is Map ? '${item['text'] ?? ''}' : '$item').join('\n');
    _answers.text = _answerIndexes(options, challenge.answers).join(', ');
  }

  List<String> _pairLines(dynamic options, List<dynamic> answers, {required bool imageMode}) {
    if (options is! Map) return const [];
    final left = options['left'] is List ? options['left'] as List : const [];
    final right = options['right'] is List ? options['right'] as List : const [];
    final rightById = {
      for (final item in right)
        if (item is Map) '${item['id']}': item,
    };
    if (answers.isEmpty) {
      return List.generate(left.length < right.length ? left.length : right.length, (index) {
        final leftMap = left[index] is Map ? left[index] as Map : const {};
        final rightMap = right[index] is Map ? right[index] as Map : const {};
        return imageMode
            ? '${leftMap['image_url'] ?? ''} | ${leftMap['alt'] ?? ''} | ${rightMap['text'] ?? ''}'
            : '${leftMap['text'] ?? ''} | ${rightMap['text'] ?? ''}';
      });
    }
    return answers.map((answer) {
      final answerMap = answer is Map ? answer : const {};
      final leftMap = left.cast<dynamic>().firstWhere(
            (item) => item is Map && '${item['id']}' == '${answerMap['left_id']}',
            orElse: () => const {},
          ) as Map;
      final rightMap = rightById['${answerMap['right_id']}'] ?? const {};
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
    return [
      for (final answer in answers)
        if (ids.indexOf('$answer') >= 0) ids.indexOf('$answer') + 1,
    ];
  }

  Map<String, dynamic> _payloadValue() {
    final type = _type.text;
    final lines = _nonEmptyLines(_payload.text);
    if (type == 'theory') {
      return {
        'facts': lines.length > 1 ? lines.take(lines.length - 1).toList() : lines,
        'summary': lines.length > 1 ? lines.last : '',
      };
    }
    if (type == 'image_question') {
      return {'image_url': lines.isNotEmpty ? lines.first : '', 'alt': lines.length > 1 ? lines.sublist(1).join(' ') : ''};
    }
    if (type == 'quote_question') {
      return {'quote': lines.isNotEmpty ? lines.first : '', 'source': lines.length > 1 ? lines.sublist(1).join(' ') : ''};
    }
    if (type == 'fill_in_blank') {
      return {'text': _payload.text.trim(), 'placeholder': '____'};
    }
    return {};
  }

  dynamic _optionsValue() {
    final type = _type.text;
    if (type == 'theory' || type == 'fill_in_blank') return [];
    if (type == 'true_false') {
      return [
        {'id': 'true', 'text': 'Верно'},
        {'id': 'false', 'text': 'Неверно'},
      ];
    }
    if (type == 'timeline') {
      return _nonEmptyLines(_options.text).asMap().entries.map((entry) {
        final parts = _splitParts(entry.value);
        return {'id': _idFor(entry.key), 'date': parts.isNotEmpty ? parts.first : '', 'text': parts.length > 1 ? parts.sublist(1).join(' | ') : entry.value};
      }).toList();
    }
    if (type == 'match_pairs' || type == 'match_image') {
      final lines = _nonEmptyLines(_options.text);
      return {
        'left': lines.asMap().entries.map((entry) {
          final parts = _splitParts(entry.value);
          return type == 'match_image'
              ? {'id': 'l${entry.key + 1}', 'image_url': parts.isNotEmpty ? parts[0] : '', 'alt': parts.length > 1 ? parts[1] : ''}
              : {'id': 'l${entry.key + 1}', 'text': parts.isNotEmpty ? parts[0] : ''};
        }).toList(),
        'right': lines.asMap().entries.map((entry) {
          final parts = _splitParts(entry.value);
          return {'id': 'r${entry.key + 1}', 'text': parts.length > (type == 'match_image' ? 2 : 1) ? parts[type == 'match_image' ? 2 : 1] : ''};
        }).toList(),
      };
    }
    return _choiceList();
  }

  List<dynamic> _answersValue() {
    final type = _type.text;
    if (type == 'theory') return [];
    if (type == 'match_pairs' || type == 'match_image') {
      final count = _nonEmptyLines(_options.text).length;
      return List.generate(count, (index) => {'left_id': 'l${index + 1}', 'right_id': 'r${index + 1}'});
    }
    if (type == 'true_false') {
      final text = _answers.text.trim().toLowerCase();
      return [text.contains('не') || text == 'false' ? 'false' : 'true'];
    }
    if (type == 'fill_in_blank') return _nonEmptyLines(_answers.text);
    final optionIds = _choiceList().map((item) => '${item['id']}').toList();
    final indexes = _parseAnswerIndexes(_answers.text, optionIds.length);
    return [for (final index in indexes) optionIds[index]];
  }

  List<Map<String, String>> _choiceList() {
    return _nonEmptyLines(_options.text).asMap().entries.map((entry) => {'id': _idFor(entry.key), 'text': entry.value}).toList();
  }

  List<String> _nonEmptyLines(String text) => text.split('\n').map((line) => line.trim()).where((line) => line.isNotEmpty).toList();

  List<String> _splitParts(String line) => line.split('|').map((part) => part.trim()).toList();

  String _idFor(int index) => String.fromCharCode('a'.codeUnitAt(0) + index);

  List<int> _parseAnswerIndexes(String text, int optionCount) {
    final parts = text.split(RegExp(r'[,;\s]+')).map((part) => part.trim().toLowerCase()).where((part) => part.isNotEmpty);
    final indexes = <int>[];
    for (final part in parts) {
      final number = int.tryParse(part);
      final index = number != null ? number - 1 : part.codeUnitAt(0) - 'a'.codeUnitAt(0);
      if (index >= 0 && index < optionCount && !indexes.contains(index)) indexes.add(index);
    }
    return indexes.isEmpty && optionCount > 0 ? [0] : indexes;
  }

  bool _usesBody(String type) => type == 'theory';
  bool _usesPayload(String type) => ['theory', 'image_question', 'quote_question', 'fill_in_blank'].contains(type);
  bool _usesOptions(String type) => !['fill_in_blank', 'theory', 'true_false'].contains(type);
  bool _usesAnswers(String type) => !['theory', 'match_pairs', 'match_image'].contains(type);
  String _promptLabel(String type) => switch (type) {
        'theory' => 'Заголовок теории',
        _ => 'Вопрос для ученика',
      };
  String _payloadLabel(String type) => switch (type) {
        'theory' => 'Факты и краткое резюме',
        'image_question' => 'Ссылка на изображение и описание',
        'quote_question' => 'Цитата и источник',
        'fill_in_blank' => 'Текст с пропуском',
        _ => 'Дополнительные данные',
      };
  String _optionsLabel(String type) => switch (type) {
        'match_pairs' || 'match_image' => 'Пары для сопоставления',
        'timeline' => 'События для сортировки',
        'true_false' => 'Варианты “верно / неверно”',
        _ => 'Варианты ответа',
      };
  String _answersLabel(String type) => switch (type) {
        'timeline' => 'Правильный порядок',
        'match_pairs' || 'match_image' => 'Правильные пары',
        'true_false' => 'Правильный выбор',
        'fill_in_blank' => 'Допустимые ответы',
        _ => 'Правильный ответ',
      };
  String _explanationLabel(String type) => switch (type) {
        'theory' => 'Что важно запомнить',
        _ => 'Объяснение при ошибке',
      };
}

class _FormDialog extends StatelessWidget {
  final String title;
  final List<Widget> fields;
  final VoidCallback onSave;
  final String? error;

  const _FormDialog({required this.title, required this.fields, required this.onSave, this.error});

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
                  Expanded(child: Text(title, style: GoogleFonts.playfairDisplay(color: AppTheme.textPrimary, fontSize: 20, fontWeight: FontWeight.bold))),
                  IconButton(onPressed: () => Navigator.pop(context), icon: Icon(Icons.close, color: AppTheme.textSecondary)),
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
                child: Text(error!, style: GoogleFonts.lato(color: AppTheme.error, fontWeight: FontWeight.bold)),
              ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(onPressed: onSave, child: const Text('Сохранить')),
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

  const _TextInput({required this.controller, required this.label, this.maxLines = 1, this.keyboardType});

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      maxLines: maxLines,
      keyboardType: keyboardType,
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
    fillColor: AppTheme.primary.withOpacity(0.35),
    enabledBorder: OutlineInputBorder(borderSide: BorderSide(color: AppTheme.cardBg), borderRadius: BorderRadius.circular(8)),
    focusedBorder: OutlineInputBorder(borderSide: const BorderSide(color: AppTheme.accent), borderRadius: BorderRadius.circular(8)),
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
      decoration: BoxDecoration(color: AppTheme.error.withOpacity(0.12), borderRadius: BorderRadius.circular(8), border: Border.all(color: AppTheme.error.withOpacity(0.4))),
      child: Text(message, style: GoogleFonts.lato(color: AppTheme.error, fontWeight: FontWeight.bold)),
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
      decoration: BoxDecoration(color: AppTheme.surface, borderRadius: BorderRadius.circular(8), border: Border.all(color: AppTheme.cardBg)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: GoogleFonts.lato(color: AppTheme.textPrimary, fontWeight: FontWeight.bold)),
          const SizedBox(height: 4),
          Text(subtitle, style: GoogleFonts.lato(color: AppTheme.textSecondary)),
        ],
      ),
    );
  }
}

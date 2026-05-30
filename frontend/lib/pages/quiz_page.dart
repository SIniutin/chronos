import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:google_fonts/google_fonts.dart';
import '../api/api_client.dart';
import '../api/content_api.dart';
import '../api/learning_api.dart';
import '../theme/app_theme.dart';
import '../data/app_data.dart';
import '../models/models.dart';
import '../state/session_controller.dart';

class QuizPage extends StatefulWidget {
  final List<QuizQuestion>? questions;
  final String? skillId;

  const QuizPage({super.key, this.questions, this.skillId});

  @override
  State<QuizPage> createState() => _QuizPageState();
}

class _QuizPageState extends State<QuizPage> with TickerProviderStateMixin {
  int _currentQuestion = 0;
  int? _selectedAnswer;
  bool _answered = false;
  int _score = 0;
  bool _finished = false;
  late AnimationController _progressController;
  final _fillController = TextEditingController();
  LearningApi? _learningApi;
  LessonSessionDto? _session;
  CurrentChallengeDto? _current;
  SubmitAnswerResultDto? _backendAnswer;
  LessonSessionResultDto? _backendResult;
  bool _backendStarted = false;
  bool _backendBusy = false;
  String? _backendError;
  String? _selectedOptionId;
  final Set<String> _selectedOptionIds = {};
  final List<String> _orderedOptionIds = [];
  final Map<String, String> _pairSelections = {};

  List<QuizQuestion> get _questions =>
      widget.questions == null || widget.questions!.isEmpty ? AppData.quizQuestions : widget.questions!;
  bool get _usesBackend => widget.skillId != null && widget.skillId!.isNotEmpty;

  @override
  void initState() {
    super.initState();
    _progressController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 500),
    );
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_usesBackend && !_backendStarted) {
      _backendStarted = true;
      _learningApi = LearningApi(SessionScope.of(context).client);
      _startBackendSession();
    }
  }

  @override
  void dispose() {
    _progressController.dispose();
    _fillController.dispose();
    super.dispose();
  }

  void _selectAnswer(int index) {
    if (_answered) return;
    setState(() {
      _selectedAnswer = index;
      _answered = true;
      if (_questions[_currentQuestion].correctIndex != null && index == _questions[_currentQuestion].correctIndex) {
        _score++;
      }
    });
  }

  void _nextQuestion() {
    if (_currentQuestion < _questions.length - 1) {
      setState(() {
        _currentQuestion++;
        _selectedAnswer = null;
        _answered = false;
      });
    } else {
      setState(() => _finished = true);
    }
  }

  Future<void> _startBackendSession() async {
    setState(() {
      _backendBusy = true;
      _backendError = null;
    });
    try {
      final session = await _learningApi!.startSession(widget.skillId!);
      final current = await _learningApi!.getCurrentChallenge(session.id);
      if (!mounted) return;
      setState(() {
        _session = session;
        _current = current;
        _backendBusy = false;
        _resetBackendAnswerState(current.challenge);
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _backendError = error.message;
        _backendBusy = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _backendError = 'Не удалось начать урок';
        _backendBusy = false;
      });
    }
  }

  Future<void> _submitBackendAnswer() async {
    final session = _session;
    final current = _current;
    if (session == null || current == null || _backendBusy || _backendAnswer != null) return;

    setState(() {
      _backendBusy = true;
      _backendError = null;
    });
    try {
      final result = await _learningApi!.submitAnswer(session.id, _answerPayload(current.challenge));
      if (!mounted) return;
      setState(() {
        _backendAnswer = result;
        _backendBusy = false;
        if (result.isCorrect) {
          _score++;
        }
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _backendError = error.message;
        _backendBusy = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _backendError = 'Не удалось отправить ответ';
        _backendBusy = false;
      });
    }
  }

  Future<void> _nextBackendChallenge() async {
    final session = _session;
    final answer = _backendAnswer;
    if (session == null || answer == null || _backendBusy) return;
    if (!answer.hasNext) {
      await _finishBackendSession();
      return;
    }

    setState(() {
      _backendBusy = true;
      _backendError = null;
    });
    try {
      final current = await _learningApi!.getCurrentChallenge(session.id);
      if (!mounted) return;
      setState(() {
        _current = current;
        _currentQuestion++;
        _backendBusy = false;
        _resetBackendAnswerState(current.challenge);
      });
    } on ApiException catch (error) {
      if (error.statusCode == 404) {
        await _finishBackendSession();
        return;
      }
      if (!mounted) return;
      setState(() {
        _backendError = error.message;
        _backendBusy = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _backendError = 'Не удалось загрузить следующий вопрос';
        _backendBusy = false;
      });
    }
  }

  Future<void> _finishBackendSession() async {
    final session = _session;
    if (session == null) return;
    setState(() {
      _backendBusy = true;
      _backendError = null;
    });
    try {
      final result = await _learningApi!.finishSession(session.id);
      if (!mounted) return;
      setState(() {
        _backendResult = result;
        _finished = true;
        _backendBusy = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _backendError = error.message;
        _backendBusy = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _backendError = 'Не удалось завершить урок';
        _backendBusy = false;
      });
    }
  }

  void _resetBackendAnswerState([ChallengeDto? challenge]) {
    _backendAnswer = null;
    _selectedOptionId = null;
    _selectedOptionIds.clear();
    _orderedOptionIds
      ..clear()
      ..addAll(_options(challenge?.options).asMap().entries.map((entry) => _optionId(entry.value, entry.key)));
    _pairSelections.clear();
    _fillController.clear();
  }

  @override
  Widget build(BuildContext context) {
    if (_usesBackend) {
      return _buildBackend(context);
    }

    if (_finished) {
      return _buildResults(context);
    }

    final question = _questions[_currentQuestion];
    final progress = (_currentQuestion + 1) / _questions.length;
    final knownAnswer = question.correctIndex != null;
    final isWrong = knownAnswer && _answered && _selectedAnswer != question.correctIndex;

    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: const Text('Исторический квиз'),
        leading: IconButton(
          icon: const Icon(Icons.close, color: AppTheme.accent),
          onPressed: () => Navigator.maybePop(context),
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Center(
              child: Text(
                '${_currentQuestion + 1}/${_questions.length}',
                style: GoogleFonts.lato(
                  color: AppTheme.textSecondary,
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Progress
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: progress,
                backgroundColor: AppTheme.surface,
                valueColor: const AlwaysStoppedAnimation<Color>(AppTheme.accent),
                minHeight: 8,
              ),
            ),

            const SizedBox(height: 28),

            // Score indicator
            Row(
              children: [
                const Text('⭐', style: TextStyle(fontSize: 16)),
                const SizedBox(width: 6),
                Text(
                  'Счёт: $_score',
                  style: GoogleFonts.lato(
                    color: AppTheme.accent,
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),

            const SizedBox(height: 24),

            // Question
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: AppTheme.surface,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: AppTheme.cardBg, width: 1),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Вопрос ${_currentQuestion + 1}',
                    style: GoogleFonts.lato(
                      color: AppTheme.accent,
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 1.5,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Text(
                    question.question,
                    style: GoogleFonts.playfairDisplay(
                      color: AppTheme.textPrimary,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      height: 1.4,
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 24),

            // Options
            Expanded(
              child: ListView.separated(
                itemCount: question.options.length,
                separatorBuilder: (_, __) => const SizedBox(height: 10),
                itemBuilder: (context, i) => _AnswerOption(
                  text: question.options[i],
                  index: i,
                  isSelected: _selectedAnswer == i,
                  isCorrect: question.correctIndex == i,
                  isAnswered: _answered,
                  showCorrectness: knownAnswer,
                  onTap: () => _selectAnswer(i),
                ),
              ),
            ),

            // Explanation + Next
            if (_answered) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: !knownAnswer
                      ? AppTheme.accent.withOpacity(0.1)
                      : _selectedAnswer == question.correctIndex
                      ? AppTheme.correct.withOpacity(0.1)
                      : AppTheme.wrong.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(
                    color: !knownAnswer
                        ? AppTheme.accent.withOpacity(0.4)
                        : _selectedAnswer == question.correctIndex
                        ? AppTheme.correct.withOpacity(0.4)
                        : AppTheme.wrong.withOpacity(0.4),
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(
                          !knownAnswer
                              ? 'Ответ принят'
                              : _selectedAnswer == question.correctIndex
                                  ? '✅ Верно!'
                                  : '❌ Неверно',
                          style: GoogleFonts.lato(
                            color: !knownAnswer
                                ? AppTheme.accent
                                : _selectedAnswer == question.correctIndex
                                ? AppTheme.correct
                                : AppTheme.wrong,
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    if (isWrong) ...[
                      const SizedBox(height: 6),
                      Text(
                        question.explanation,
                        style: GoogleFonts.lato(
                          color: AppTheme.textPrimary,
                          fontSize: 13,
                          height: 1.5,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _nextQuestion,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppTheme.accent,
                    foregroundColor: AppTheme.onAccent,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  ),
                  child: Text(
                    _currentQuestion < _questions.length - 1
                        ? 'Следующий вопрос →'
                        : 'Узнать результат 🏆',
                    style: GoogleFonts.lato(fontSize: 15, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ],
            const SizedBox(height: 10),
          ],
        ),
      ),
    );
  }

  Widget _buildBackend(BuildContext context) {
    if (_finished) {
      return _buildResults(context);
    }

    final current = _current;
    final challenge = current?.challenge;
    final total = _backendResult?.total ?? 10;
    final progress = total <= 0 ? 0.0 : ((_currentQuestion + 1) / total).clamp(0.0, 1.0).toDouble();
    final canSubmit = challenge != null && _canSubmitBackendAnswer(challenge);

    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: const Text('Урок'),
        leading: IconButton(
          icon: const Icon(Icons.close, color: AppTheme.accent),
          onPressed: () => Navigator.maybePop(context),
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Center(
              child: Text(
                '${_currentQuestion + 1}',
                style: GoogleFonts.lato(
                  color: AppTheme.textSecondary,
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: _backendBusy && challenge == null
              ? const Center(child: CircularProgressIndicator(color: AppTheme.accent))
              : _backendError != null && challenge == null
                  ? _ErrorState(message: _backendError!, onRetry: _startBackendSession)
                  : challenge == null
                      ? _ErrorState(message: 'В уроке пока нет заданий', onRetry: _startBackendSession)
                      : Column(
                          children: [
                            Expanded(
                              child: SingleChildScrollView(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    ClipRRect(
                                      borderRadius: BorderRadius.circular(4),
                                      child: LinearProgressIndicator(
                                        value: progress,
                                        backgroundColor: AppTheme.surface,
                                        valueColor: const AlwaysStoppedAnimation<Color>(AppTheme.accent),
                                        minHeight: 8,
                                      ),
                                    ),
                                    const SizedBox(height: 24),
                                    _BackendChallengeCard(challenge: challenge),
                                    const SizedBox(height: 18),
                                    _buildBackendAnswerInput(challenge),
                                    if (_backendError != null) ...[
                                      const SizedBox(height: 14),
                                      Text(
                                        _backendError!,
                                        style: GoogleFonts.lato(color: AppTheme.wrong, fontSize: 13),
                                      ),
                                    ],
                                    if (_backendAnswer != null) ...[
                                      const SizedBox(height: 14),
                                      _BackendFeedback(result: _backendAnswer!),
                                    ],
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 14),
                            SizedBox(
                              width: double.infinity,
                              child: ElevatedButton(
                                onPressed: _backendBusy
                                    ? null
                                    : _backendAnswer == null
                                        ? canSubmit
                                            ? _submitBackendAnswer
                                            : null
                                        : _nextBackendChallenge,
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: AppTheme.accent,
                                  foregroundColor: AppTheme.onAccent,
                                  padding: const EdgeInsets.symmetric(vertical: 14),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                                ),
                                child: Text(
                                  _backendAnswer == null
                                      ? 'Ответить'
                                      : _backendAnswer!.hasNext
                                          ? 'Дальше'
                                          : 'Завершить',
                                  style: GoogleFonts.lato(fontSize: 15, fontWeight: FontWeight.bold),
                                ),
                              ),
                            ),
                          ],
                        ),
        ),
      ),
    );
  }

  Widget _buildBackendAnswerInput(ChallengeDto challenge) {
    if (challenge.type == 'theory') {
      return Center(
        child: Text(
          'Когда будешь готов, продолжай к заданиям.',
          style: GoogleFonts.lato(color: AppTheme.textPrimary, fontSize: 15, height: 1.65),
          textAlign: TextAlign.center,
        ),
      );
    }

    if (challenge.type == 'fill_in_blank') {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          TextField(
            controller: _fillController,
            enabled: _backendAnswer == null,
            onChanged: (_) => setState(() {}),
            style: GoogleFonts.lato(color: AppTheme.textPrimary),
            decoration: InputDecoration(
              hintText: 'Введите ответ',
              hintStyle: GoogleFonts.lato(color: AppTheme.textSecondary),
              filled: true,
              fillColor: AppTheme.surface,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: BorderSide(color: AppTheme.cardBg),
              ),
            ),
          ),
        ],
      );
    }

    if (challenge.type == 'match_pairs' || challenge.type == 'match_image') {
      return _buildMatchingInput(challenge);
    }

    if (challenge.type == 'timeline') {
      return _buildTimelineInput(challenge);
    }

    final options = _options(challenge.options);
    if (options.isEmpty) {
      return _UnsupportedChallenge(message: 'У задания нет вариантов ответа.');
    }

    final multi = challenge.type == 'multiple_choice';
    return Column(
      children: [
        for (final entry in options.asMap().entries) ...[
          Builder(
            builder: (context) {
              final option = entry.value;
              final id = _optionId(option, entry.key);
              final selected = multi ? _selectedOptionIds.contains(id) : _selectedOptionId == id;
              return _AnswerOption(
                text: _optionText(option),
                index: entry.key,
                isSelected: selected,
                isCorrect: false,
                isAnswered: _backendAnswer != null,
                showCorrectness: false,
                onTap: () {
                  if (_backendAnswer != null) return;
                  setState(() {
                    if (multi) {
                      if (_selectedOptionIds.contains(id)) {
                        _selectedOptionIds.remove(id);
                      } else {
                        _selectedOptionIds.add(id);
                      }
                    } else {
                      _selectedOptionId = id;
                    }
                  });
                },
              );
            },
          ),
          if (entry.key != options.length - 1) const SizedBox(height: 10),
        ],
      ],
    );
  }

  Widget _buildTimelineInput(ChallengeDto challenge) {
    final options = _options(challenge.options);
    if (options.isEmpty) {
      return _UnsupportedChallenge(message: 'Для задания на порядок нет элементов.');
    }
    if (_orderedOptionIds.length != options.length) {
      _orderedOptionIds
        ..clear()
        ..addAll(options.asMap().entries.map((entry) => _optionId(entry.value, entry.key)));
    }
    final byId = {
      for (final entry in options.asMap().entries) _optionId(entry.value, entry.key): entry.value,
    };
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Расставь элементы в правильном порядке',
          style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 13),
        ),
        const SizedBox(height: 10),
        ReorderableListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: _orderedOptionIds.length,
          onReorder: _backendAnswer != null
              ? (_, __) {}
              : (oldIndex, newIndex) {
                  setState(() {
                    if (newIndex > oldIndex) newIndex--;
                    final id = _orderedOptionIds.removeAt(oldIndex);
                    _orderedOptionIds.insert(newIndex, id);
                  });
                },
          itemBuilder: (context, index) {
            final id = _orderedOptionIds[index];
            final option = byId[id];
            return Padding(
              key: ValueKey(id),
              padding: const EdgeInsets.only(bottom: 10),
              child: _AnswerOption(
                text: _optionText(option),
                index: index,
                isSelected: true,
                isCorrect: false,
                isAnswered: _backendAnswer != null,
                showCorrectness: false,
                onTap: () {},
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _buildMatchingInput(ChallengeDto challenge) {
    final groups = _matchingOptions(challenge.options);
    final left = groups.left;
    final right = groups.right;
    if (left.isEmpty || right.isEmpty) {
      return _UnsupportedChallenge(message: 'Для задания на сопоставление не хватает пар.');
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Выбери соответствие для каждого пункта',
          style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 13),
        ),
        const SizedBox(height: 12),
        for (final entry in left.asMap().entries) ...[
          _MatchRow(
            leftText: _optionText(entry.value),
            value: _pairSelections[_optionId(entry.value, entry.key)],
            rightOptions: right.asMap().entries
                .map((rightEntry) => _MatchChoice(
                      id: _optionId(rightEntry.value, rightEntry.key),
                      text: _optionText(rightEntry.value),
                    ))
                .toList(),
            enabled: _backendAnswer == null,
            onChanged: (rightId) {
              if (rightId == null) return;
              setState(() => _pairSelections[_optionId(entry.value, entry.key)] = rightId);
            },
          ),
          if (entry.key != left.length - 1) const SizedBox(height: 12),
        ],
      ],
    );
  }

  bool _canSubmitBackendAnswer(ChallengeDto challenge) {
    if (challenge.type == 'theory') return true;
    if (challenge.type == 'fill_in_blank') return _fillController.text.trim().isNotEmpty;
    if (challenge.type == 'multiple_choice') return _selectedOptionIds.isNotEmpty;
    if (challenge.type == 'timeline') return _orderedOptionIds.isNotEmpty && _orderedOptionIds.length == _options(challenge.options).length;
    if (challenge.type == 'match_pairs' || challenge.type == 'match_image') {
      final groups = _matchingOptions(challenge.options);
      return groups.left.isNotEmpty && _pairSelections.length == groups.left.length;
    }
    if (_isSingleAnswerType(challenge.type)) return _selectedOptionId != null;
    return false;
  }

  bool _isSingleAnswerType(String type) {
    return type == 'single_choice' ||
        type == 'true_false' ||
        type == 'image_question' ||
        type == 'quote_question';
  }

  Object? _answerPayload(ChallengeDto challenge) {
    if (challenge.type == 'theory') {
      return true;
    }
    if (challenge.type == 'fill_in_blank') {
      return _fillController.text.trim();
    }
    if (challenge.type == 'multiple_choice') {
      return _selectedOptionIds.toList();
    }
    if (challenge.type == 'timeline') {
      return _orderedOptionIds.toList();
    }
    if (challenge.type == 'match_pairs' || challenge.type == 'match_image') {
      final left = _matchingOptions(challenge.options).left;
      return left.asMap().entries
          .map((entry) => {
                'left_id': _optionId(entry.value, entry.key),
                'right_id': _pairSelections[_optionId(entry.value, entry.key)],
              })
          .toList();
    }
    return _selectedOptionId;
  }

  List<dynamic> _options(dynamic raw) {
    if (raw is List) return raw;
    return const [];
  }

  _MatchingOptions _matchingOptions(dynamic raw) {
    if (raw is Map) {
      return _MatchingOptions(_options(raw['left']), _options(raw['right']));
    }
    return const _MatchingOptions([], []);
  }

  String _optionId(dynamic option, int index) {
    if (option is Map && option['id'] != null) return option['id'].toString();
    if (option is Map && option['value'] != null) return option['value'].toString();
    return index.toString();
  }

  String _optionText(dynamic option) {
    if (option is Map && option['text'] != null) return option['text'].toString();
    if (option is Map && option['label'] != null) return option['label'].toString();
    if (option is Map && option['value'] != null) return option['value'].toString();
    return option.toString();
  }

  Widget _buildResults(BuildContext context) {
    final total = _backendResult?.total ?? _questions.length;
    final score = _backendResult?.correct ?? _score;
    final percent = _backendResult?.percent ?? (total == 0 ? 0 : (score / total * 100).round());
    String emoji;
    String title;
    String subtitle;

    if (percent >= 80) {
      emoji = '🏆';
      title = 'Блестяще!';
      subtitle = 'Ты настоящий знаток истории';
    } else if (percent >= 60) {
      emoji = '👍';
      title = 'Хорошо!';
      subtitle = 'Ещё немного и будешь историком';
    } else {
      emoji = '📚';
      title = 'Нужно подтянуться';
      subtitle = 'Почитай уроки и попробуй снова';
    }

    return Scaffold(
      backgroundColor: AppTheme.primary,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(emoji, style: const TextStyle(fontSize: 80)),
              const SizedBox(height: 20),
              Text(
                title,
                style: GoogleFonts.playfairDisplay(
                  color: AppTheme.textPrimary,
                  fontSize: 32,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                subtitle,
                style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 16),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 40),
              Container(
                padding: const EdgeInsets.all(30),
                decoration: BoxDecoration(
                  color: AppTheme.surface,
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(color: AppTheme.accent.withOpacity(0.3), width: 1.5),
                ),
                child: Column(
                  children: [
                    Text(
                      '$score / $total',
                      style: GoogleFonts.playfairDisplay(
                        color: AppTheme.accent,
                        fontSize: 52,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    Text(
                      'правильных ответов',
                      style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 14),
                    ),
                    const SizedBox(height: 16),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(6),
                      child: LinearProgressIndicator(
                        value: total == 0 ? 0.0 : score / total,
                        backgroundColor: AppTheme.cardBg,
                        valueColor: AlwaysStoppedAnimation<Color>(
                          percent >= 80 ? AppTheme.correct : percent >= 60 ? AppTheme.accent : AppTheme.wrong,
                        ),
                        minHeight: 10,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Text(
                      '$percent%',
                      style: GoogleFonts.lato(
                        color: AppTheme.accent,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 30),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => Navigator.pop(context),
                      style: OutlinedButton.styleFrom(
                        side: const BorderSide(color: AppTheme.accent),
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                      ),
                      child: Text('На главную', style: GoogleFonts.lato(color: AppTheme.accent, fontSize: 15)),
                    ),
                  ),
                  if (!_usesBackend) ...[
                    const SizedBox(width: 12),
                    Expanded(
                      child: ElevatedButton(
                        onPressed: () {
                          setState(() {
                            _currentQuestion = 0;
                            _selectedAnswer = null;
                            _answered = false;
                            _score = 0;
                            _finished = false;
                          });
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppTheme.accent,
                          foregroundColor: AppTheme.onAccent,
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                        ),
                        child: Text('Ещё раз', style: GoogleFonts.lato(fontSize: 15, fontWeight: FontWeight.bold)),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _BackendChallengeCard extends StatelessWidget {
  final ChallengeDto challenge;

  const _BackendChallengeCard({required this.challenge});

  @override
  Widget build(BuildContext context) {
    final body = challenge.type == 'fill_in_blank' ? _payloadText(challenge.payload) : _bodyText();
    final imageUrl = _imageUrl(challenge.payload);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(22),
      decoration: BoxDecoration(
        color: AppTheme.surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: AppTheme.cardBg, width: 1),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            _typeLabel(challenge.type),
            style: GoogleFonts.lato(
              color: AppTheme.accent,
              fontSize: 12,
              fontWeight: FontWeight.bold,
            ),
          ),
          if (imageUrl != null) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: CachedNetworkImage(
                imageUrl: imageUrl,
                fit: BoxFit.cover,
                width: double.infinity,
                height: 180,
                placeholder: (_, __) => Container(
                  height: 180,
                  color: AppTheme.cardBg,
                  child: const Center(child: CircularProgressIndicator(color: AppTheme.accent)),
                ),
                errorWidget: (_, __, ___) => Container(
                  height: 180,
                  color: AppTheme.cardBg,
                  alignment: Alignment.center,
                  child: Icon(Icons.broken_image_outlined, color: AppTheme.textSecondary),
                ),
              ),
            ),
          ],
          const SizedBox(height: 10),
          Text(
            challenge.prompt,
            style: GoogleFonts.playfairDisplay(
              color: AppTheme.textPrimary,
              fontSize: 20,
              fontWeight: FontWeight.bold,
              height: 1.35,
            ),
          ),
          if (body.trim().isNotEmpty) ...[
            const SizedBox(height: 12),
            Text(
              body,
              style: GoogleFonts.lato(
                color: AppTheme.textPrimary,
                fontSize: 14,
                height: 1.55,
              ),
            ),
          ],
        ],
      ),
    );
  }

  String _payloadText(dynamic payload) {
    if (payload is Map && payload['text'] != null) {
      return payload['text'].toString();
    }
    return challenge.body;
  }

  String _bodyText() {
    if (challenge.body.trim().isNotEmpty) {
      return challenge.body;
    }
    if (challenge.payload is Map) {
      final payload = challenge.payload as Map;
      if (payload['quote'] != null) return payload['quote'].toString();
      if (payload['text'] != null) return payload['text'].toString();
      if (payload['caption'] != null) return payload['caption'].toString();
    }
    return '';
  }

  String? _imageUrl(dynamic payload) {
    if (payload is! Map) return null;
    final raw = payload['image_url'] ?? payload['imageUrl'] ?? payload['image'] ?? payload['url'];
    if (raw == null || raw.toString().trim().isEmpty) return null;
    return raw.toString();
  }

  String _typeLabel(String type) => switch (type) {
        'theory' => 'ТЕОРИЯ',
        'single_choice' => 'ОДИН ОТВЕТ',
        'true_false' => 'ВЕРНО ИЛИ НЕТ',
        'fill_in_blank' => 'ЗАПОЛНИ ПРОПУСК',
        'multiple_choice' => 'НЕСКОЛЬКО ОТВЕТОВ',
        'timeline' => 'ПОРЯДОК',
        'match_pairs' => 'СОПОСТАВЛЕНИЕ',
        'match_image' => 'СОПОСТАВЬ ИЗОБРАЖЕНИЯ',
        'image_question' => 'ВОПРОС ПО ИЗОБРАЖЕНИЮ',
        'quote_question' => 'ВОПРОС ПО ЦИТАТЕ',
        _ => 'ЗАДАНИЕ',
      };
}

class _MatchingOptions {
  final List<dynamic> left;
  final List<dynamic> right;

  const _MatchingOptions(this.left, this.right);
}

class _MatchChoice {
  final String id;
  final String text;

  const _MatchChoice({required this.id, required this.text});
}

class _MatchRow extends StatelessWidget {
  final String leftText;
  final String? value;
  final List<_MatchChoice> rightOptions;
  final bool enabled;
  final ValueChanged<String?> onChanged;

  const _MatchRow({
    required this.leftText,
    required this.value,
    required this.rightOptions,
    required this.enabled,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppTheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppTheme.cardBg),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            leftText,
            style: GoogleFonts.lato(color: AppTheme.textPrimary, fontSize: 14, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 10),
          DropdownButtonFormField<String>(
            value: value,
            isExpanded: true,
            dropdownColor: AppTheme.surface,
            decoration: InputDecoration(
              labelText: 'Соответствие',
              labelStyle: GoogleFonts.lato(color: AppTheme.textSecondary),
              filled: true,
              fillColor: AppTheme.primary.withOpacity(0.35),
              enabledBorder: OutlineInputBorder(
                borderSide: BorderSide(color: AppTheme.cardBg),
                borderRadius: BorderRadius.circular(10),
              ),
              focusedBorder: OutlineInputBorder(
                borderSide: const BorderSide(color: AppTheme.accent),
                borderRadius: BorderRadius.circular(10),
              ),
            ),
            items: rightOptions
                .map(
                  (choice) => DropdownMenuItem(
                    value: choice.id,
                    child: Text(
                      choice.text,
                      overflow: TextOverflow.ellipsis,
                      style: GoogleFonts.lato(color: AppTheme.textPrimary),
                    ),
                  ),
                )
                .toList(),
            onChanged: enabled ? onChanged : null,
          ),
        ],
      ),
    );
  }
}

class _UnsupportedChallenge extends StatelessWidget {
  final String message;

  const _UnsupportedChallenge({required this.message});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.wrong.withOpacity(0.08),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppTheme.wrong.withOpacity(0.35)),
      ),
      child: Text(
        message,
        style: GoogleFonts.lato(color: AppTheme.wrong, fontSize: 14, fontWeight: FontWeight.bold),
        textAlign: TextAlign.center,
      ),
    );
  }
}

class _BackendFeedback extends StatelessWidget {
  final SubmitAnswerResultDto result;

  const _BackendFeedback({required this.result});

  @override
  Widget build(BuildContext context) {
    final color = result.isCorrect ? AppTheme.correct : AppTheme.wrong;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        result.isCorrect ? 'Верно' : 'Ответ принят, но есть ошибка',
        style: GoogleFonts.lato(
          color: color,
          fontSize: 14,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorState({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.error_outline, color: AppTheme.wrong, size: 34),
          const SizedBox(height: 12),
          Text(
            message,
            style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 14),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          OutlinedButton(
            onPressed: onRetry,
            style: OutlinedButton.styleFrom(side: const BorderSide(color: AppTheme.accent)),
            child: Text('Повторить', style: GoogleFonts.lato(color: AppTheme.accent)),
          ),
        ],
      ),
    );
  }
}

class _AnswerOption extends StatelessWidget {
  final String text;
  final int index;
  final bool isSelected;
  final bool isCorrect;
  final bool isAnswered;
  final bool showCorrectness;
  final VoidCallback onTap;

  const _AnswerOption({
    required this.text,
    required this.index,
    required this.isSelected,
    required this.isCorrect,
    required this.isAnswered,
    required this.showCorrectness,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    Color borderColor = AppTheme.cardBg;
    Color bgColor = AppTheme.surface;
    Color textColor = AppTheme.textPrimary;
    IconData? trailingIcon;
    Color? trailingColor;

    if (isAnswered) {
      if (!showCorrectness && isSelected) {
        borderColor = AppTheme.accent;
        bgColor = AppTheme.accent.withOpacity(0.1);
        textColor = AppTheme.accent;
        trailingIcon = Icons.radio_button_checked;
        trailingColor = AppTheme.accent;
      } else if (showCorrectness && isCorrect) {
        borderColor = AppTheme.correct;
        bgColor = AppTheme.correct.withOpacity(0.1);
        textColor = AppTheme.correct;
        trailingIcon = Icons.check_circle;
        trailingColor = AppTheme.correct;
      } else if (showCorrectness && isSelected && !isCorrect) {
        borderColor = AppTheme.wrong;
        bgColor = AppTheme.wrong.withOpacity(0.1);
        textColor = AppTheme.wrong;
        trailingIcon = Icons.cancel;
        trailingColor = AppTheme.wrong;
      }
    } else if (isSelected) {
      borderColor = AppTheme.accent;
      bgColor = AppTheme.accent.withOpacity(0.1);
    }

    final letters = ['А', 'Б', 'В', 'Г'];

    return GestureDetector(
      onTap: isAnswered ? null : onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: borderColor, width: 1.5),
        ),
        child: Row(
          children: [
            Container(
              width: 34,
              height: 34,
              decoration: BoxDecoration(
                color: isAnswered && !showCorrectness && isSelected
                    ? AppTheme.accent.withOpacity(0.2)
                    : isAnswered && showCorrectness && isCorrect
                    ? AppTheme.correct.withOpacity(0.2)
                    : isAnswered && showCorrectness && isSelected && !isCorrect
                        ? AppTheme.wrong.withOpacity(0.2)
                        : AppTheme.cardBg,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Center(
                child: Text(
                  index < letters.length ? letters[index] : '${index + 1}',
                  style: GoogleFonts.lato(
                    color: textColor,
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                text,
                style: GoogleFonts.lato(
                  color: textColor,
                  fontSize: 15,
                  height: 1.4,
                ),
              ),
            ),
            if (trailingIcon != null) ...[
              const SizedBox(width: 8),
              Icon(trailingIcon, color: trailingColor, size: 20),
            ],
          ],
        ),
      ),
    );
  }
}

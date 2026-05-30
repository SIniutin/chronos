import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../widgets/history_timeline.dart';

class TimelinePage extends StatelessWidget {
  const TimelinePage({super.key});

  static const _events = [
    TimelineEvent(
      year: '1882',
      title: 'Фабричная инспекция',
      description: 'Александр III учредил Фабричную инспекцию для надзора за соблюдением рабочего законодательства на предприятиях.',
      emoji: '🏭',
      color: Color(0xFF5C7AEA),
    ),
    TimelineEvent(
      year: 'Февраль 1892',
      title: 'Витте — министр путей сообщения',
      description: 'Сергей Юльевич Витте назначен министром путей сообщения — начало его восхождения к вершинам власти.',
      emoji: '🚂',
      color: Color(0xFFE8A838),
    ),
    TimelineEvent(
      year: 'Август 1892',
      title: 'Витте возглавляет Минфин',
      description: 'Витте назначен министром финансов. Начинается эпоха масштабных экономических реформ в России.',
      emoji: '💰',
      color: Color(0xFFE8A838),
    ),
    TimelineEvent(
      year: '1894',
      title: 'Казённая винная монополия',
      description: 'Введена государственная монополия на продажу алкоголя — важнейший источник пополнения казны.',
      emoji: '🍷',
      color: Color(0xFFE8A838),
    ),
    TimelineEvent(
      year: '1894',
      title: 'Воцарение Николая II',
      description: 'После смерти Александра III на престол вступает Николай II. Россия входит в полосу бурного экономического и культурного развития.',
      emoji: '👑',
      color: Color(0xFF9B59B6),
    ),
    TimelineEvent(
      year: '1895',
      title: 'Николай II отвергает земства',
      description: 'Николай II публично отверг идеи участия земств в государственном управлении, считая самодержавие незыблемой основой.',
      emoji: '✋',
      color: Color(0xFF9B59B6),
    ),
    TimelineEvent(
      year: '1894–1904',
      title: 'Таможенные договоры с Германией',
      description: 'Заключены таможенные договоры с Германией, регулирующие торговые отношения двух крупнейших держав.',
      emoji: '🤝',
      color: Color(0xFFE8A838),
    ),
    TimelineEvent(
      year: 'Середина 1890-х',
      title: 'Введение золотого рубля',
      description: 'Денежная реформа Витте: введён золотой рубль, стабилизировавший финансовую систему и привлёкший иностранные инвестиции.',
      emoji: '🪙',
      color: Color(0xFFE8A838),
    ),
    TimelineEvent(
      year: '1903',
      title: 'Закон о возмещении ущерба рабочим',
      description: 'Введён закон, обязавший хозяев предприятий вознаграждать рабочих за каждый несчастный случай на производстве.',
      emoji: '⚖️',
      color: Color(0xFF2ECC71),
    ),
    TimelineEvent(
      year: 'Июль 1904',
      title: 'Убийство министра Плеве',
      description: 'Министр внутренних дел В. К. Плеве убит террористическим актом — начало резкого обострения внутриполитической обстановки.',
      emoji: '💥',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: '8 января 1905',
      title: 'Войска вводятся в столицу',
      description: 'Власти ввели войска в Петербург в связи с ожидавшейся радикализацией общественной обстановки.',
      emoji: '⚔️',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: '9 января 1905',
      title: 'Кровавое воскресенье',
      description: 'Войска открыли огонь по мирным рабочим колоннам, шедшим к Зимнему дворцу. Убито 96, ранено 333 человека.',
      emoji: '🩸',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: '19 января 1905',
      title: 'Николай II принимает депутацию',
      description: 'Николай II принял рабочую депутацию и заявил, что обращение к нему мятежной толпой является преступным. Аудиенция не изменила настроений.',
      emoji: '🏛️',
      color: Color(0xFF9B59B6),
    ),
    TimelineEvent(
      year: 'Зима–весна 1905',
      title: 'Беспорядки в деревне',
      description: 'Революционные волнения перекинулись на деревню: захваты и поджоги дворянских усадеб, брожение в воинских подразделениях.',
      emoji: '🔥',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: '14 июня 1905',
      title: 'Восстание на «Потёмкине»',
      description: 'Восстание команды броненосца «Князь Потёмкин Таврический». Продолжалось до 25 июня, завершилось сдачей корабля румынским властям в Констанце.',
      emoji: '⚓',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: '19 сентября 1905',
      title: 'Забастовка московских печатников',
      description: 'В Москве началась забастовка печатников с экономическими требованиями — начало процесса, приведшего к всероссийской стачке.',
      emoji: '📰',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: 'Сентябрь–октябрь 1905',
      title: 'Всеобщая стачка',
      description: 'По России развернулась всеобщая забастовка с политическими требованиями. Государственная власть столкнулась с кризисом управляемости.',
      emoji: '✊',
      color: Color(0xFFE74C3C),
    ),
    TimelineEvent(
      year: '9 октября 1905',
      title: 'Программа Витте',
      description: 'Витте представил императору программу: гражданские свободы, народное представительство, объединённый Совет министров, нормированный рабочий день.',
      emoji: '📜',
      color: Color(0xFF9B59B6),
    ),
    TimelineEvent(
      year: '17 октября 1905',
      title: 'Манифест 17 октября',
      description: 'Николай II подписал Манифест, даровавший населению гражданские свободы и учредивший законодательную Государственную думу.',
      emoji: '🕊️',
      color: Color(0xFF9B59B6),
    ),
    TimelineEvent(
      year: '1906',
      title: 'Легализация профсоюзов',
      description: 'Государство разрешило свободную организацию и деятельность профессиональных союзов рабочих и служащих.',
      emoji: '🤲',
      color: Color(0xFF2ECC71),
    ),
    TimelineEvent(
      year: '1912',
      title: 'Закон о страховании рабочих',
      description: 'Принят Закон о страховании рабочих: право на бесплатную медицинскую помощь для всех застрахованных.',
      emoji: '🏥',
      color: Color(0xFF2ECC71),
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.primary,
      appBar: AppBar(
        title: const Text('Лента истории'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: AppTheme.accent),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [AppTheme.accent.withOpacity(0.15), AppTheme.cardBg],
                ),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Row(
                children: [
                  const Text('⏳', style: TextStyle(fontSize: 28)),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      'Россия на рубеже XIX–XX вв.: ключевые события',
                      style: GoogleFonts.playfairDisplay(
                        color: AppTheme.textPrimary,
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 12),

            // Легенда
            Wrap(
              spacing: 10,
              runSpacing: 6,
              children: [
                _LegendChip(color: const Color(0xFFE8A838), label: 'Реформы Витте'),
                _LegendChip(color: const Color(0xFF9B59B6), label: 'Николай II'),
                _LegendChip(color: const Color(0xFFE74C3C), label: 'Революция 1905'),
                _LegendChip(color: const Color(0xFF2ECC71), label: 'Рабочее законодательство'),
                _LegendChip(color: const Color(0xFF5C7AEA), label: 'Фабричный надзор'),
              ],
            ),

            const SizedBox(height: 28),
            const HistoryTimeline(events: _events),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }
}

class _LegendChip extends StatelessWidget {
  final Color color;
  final String label;
  const _LegendChip({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 10,
          height: 10,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 5),
        Text(
          label,
          style: GoogleFonts.lato(color: AppTheme.textSecondary, fontSize: 11),
        ),
      ],
    );
  }
}

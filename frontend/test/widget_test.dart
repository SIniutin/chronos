import 'package:flutter_test/flutter_test.dart';

import 'package:history_app/main.dart';

void main() {
  testWidgets('History app starts', (WidgetTester tester) async {
    await tester.pumpWidget(const HistoryApp());
    await tester.pump(const Duration(seconds: 3));
    await tester.tap(find.text('Пропустить'));
    await tester.pumpAndSettle();

    expect(find.text('История'), findsWidgets);
    expect(find.text('Войти'), findsWidgets);
  });
}

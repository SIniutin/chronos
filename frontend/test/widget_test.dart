import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:history_app/main.dart';

void main() {
  testWidgets('History app starts', (WidgetTester tester) async {
    await tester.pumpWidget(const HistoryApp());
    await tester.pump();

    expect(find.byType(MaterialApp), findsOneWidget);
  });
}

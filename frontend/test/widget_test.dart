import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:radiopx/features/voice/widgets/ptt_button.dart';

void main() {
  testWidgets('PTT button renders and fires press/release callbacks',
      (WidgetTester tester) async {
    int pressed = 0;
    int released = 0;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: PttButton(
              isTransmitting: false,
              isChannelBusy: false,
              onPressed: () => pressed++,
              onReleased: () => released++,
            ),
          ),
        ),
      ),
    );

    expect(find.text('PTT'), findsOneWidget);

    final gesture = await tester.startGesture(tester.getCenter(find.byType(PttButton)));
    await tester.pump();
    await gesture.up();
    await tester.pump();

    expect(pressed, 1);
    expect(released, 1);
  });
}
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/auth/screens/login_screen.dart';

void main() {
  testWidgets('shows validation errors when submitting empty fields',
      (tester) async {
    await tester.pumpWidget(const MaterialApp(home: LoginScreen()));

    await tester.tap(find.text('Entrar'));
    await tester.pump();

    expect(find.text('Por favor, insira seu email'), findsOneWidget);
    expect(find.text('Por favor, insira sua senha'), findsOneWidget);
  });

  testWidgets('rejects a password shorter than 8 characters', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: LoginScreen()));

    await tester.enterText(
      find.byType(TextFormField).at(0),
      'user@example.com',
    );
    await tester.enterText(find.byType(TextFormField).at(1), 'short');
    await tester.tap(find.text('Entrar'));
    await tester.pump();

    expect(
      find.text('A senha deve ter pelo menos 8 caracteres'),
      findsOneWidget,
    );
  });

  testWidgets('rejects an email without @', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: LoginScreen()));

    await tester.enterText(find.byType(TextFormField).at(0), 'not-an-email');
    await tester.enterText(find.byType(TextFormField).at(1), 'password123');
    await tester.tap(find.text('Entrar'));
    await tester.pump();

    expect(find.text('Por favor, insira um email válido'), findsOneWidget);
  });
}

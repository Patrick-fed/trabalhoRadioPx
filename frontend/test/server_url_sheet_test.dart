import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/settings/widgets/server_url_sheet.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  Future<void> openSheet(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Builder(
            builder: (context) => ElevatedButton(
              onPressed: () => ServerUrlSheet.show(context),
              child: const Text('abrir'),
            ),
          ),
        ),
      ),
    );
    await tester.tap(find.text('abrir'));
    await tester.pumpAndSettle();
  }

  testWidgets('mostra a URL atual e salva URL válida', (tester) async {
    SharedPreferences.setMockInitialValues({});
    await openSheet(tester);

    expect(find.text('Servidor'), findsOneWidget);
    expect(find.text('Atual: localhost:8080'), findsOneWidget);
    expect(find.text('http://localhost:8080'), findsOneWidget);

    await tester.enterText(
        find.byType(TextField), 'https://novo-servidor.test/');
    await tester.tap(find.text('Salvar'));
    await tester.pumpAndSettle();

    final prefs = await SharedPreferences.getInstance();
    expect(prefs.getString('server_url'), 'https://novo-servidor.test');
    expect(find.text('Servidor: novo-servidor.test'), findsOneWidget);
    expect(find.byType(TextField), findsNothing);
  });

  testWidgets('recusa URL inválida com mensagem', (tester) async {
    SharedPreferences.setMockInitialValues({});
    await openSheet(tester);

    await tester.enterText(find.byType(TextField), 'ftp://errado');
    await tester.tap(find.text('Salvar'));
    await tester.pump();

    expect(
      find.text('URL inválida. Use o formato http://host ou https://host.'),
      findsOneWidget,
    );
    expect(find.byType(TextField), findsOneWidget);

    final prefs = await SharedPreferences.getInstance();
    expect(prefs.getString('server_url'), isNull);
  });
}

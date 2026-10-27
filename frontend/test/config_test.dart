import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/core/config.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
    AppConfig.resetForTest();
  });

  group('normalizeUrl', () {
    test('aceita http:// e https://', () {
      expect(AppConfig.normalizeUrl('http://localhost:8080'),
          'http://localhost:8080');
      expect(AppConfig.normalizeUrl('https://a.ngrok-free.dev'),
          'https://a.ngrok-free.dev');
    });

    test('faz trim e remove barras finais', () {
      expect(AppConfig.normalizeUrl('  https://a.ngrok-free.dev/  '),
          'https://a.ngrok-free.dev');
      expect(
          AppConfig.normalizeUrl('https://a.ngrok-free.dev///'),
          'https://a.ngrok-free.dev');
    });

    test('rejeita esquema inválido e vazio', () {
      expect(AppConfig.normalizeUrl('ftp://x'), isNull);
      expect(AppConfig.normalizeUrl('a.ngrok-free.dev'), isNull);
      expect(AppConfig.normalizeUrl(''), isNull);
      expect(AppConfig.normalizeUrl('   '), isNull);
    });
  });

  group('save/load', () {
    test('save persiste e aplica em memória', () async {
      await AppConfig.save('  https://abc.ngrok-free.dev/  ');
      expect(AppConfig.apiBaseUrl, 'https://abc.ngrok-free.dev');
      final prefs = await SharedPreferences.getInstance();
      expect(prefs.getString('server_url'), 'https://abc.ngrok-free.dev');
    });

    test('save com URL inválida lança ArgumentError', () async {
      expect(() => AppConfig.save('sem-esquema'), throwsArgumentError);
      expect(AppConfig.apiBaseUrl, 'http://localhost:8080');
    });

    test('load usa o valor salvo', () async {
      SharedPreferences.setMockInitialValues(
          {'server_url': 'https://salvo.ngrok-free.dev'});
      await AppConfig.load();
      expect(AppConfig.apiBaseUrl, 'https://salvo.ngrok-free.dev');
    });

    test('load sem valor mantém a semente do build', () async {
      await AppConfig.load();
      expect(AppConfig.apiBaseUrl, AppConfig.buildDefault);
    });
  });

  group('derivações', () {
    test('wsBaseUrl converte http→ws e https→wss', () async {
      await AppConfig.save('http://localhost:8080');
      expect(AppConfig.wsBaseUrl, 'ws://localhost:8080');
      await AppConfig.save('https://a.ngrok-free.dev');
      expect(AppConfig.wsBaseUrl, 'wss://a.ngrok-free.dev');
    });

    test('displayHost remove o esquema', () async {
      await AppConfig.save('https://a.ngrok-free.dev');
      expect(AppConfig.displayHost, 'a.ngrok-free.dev');
    });
  });
}

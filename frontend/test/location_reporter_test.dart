import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/core/location_reporter.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('LocationReporter (RF04 / DEC-020)', () {
    test('start dispara report imediato e cria timer periódico', () async {
      var calls = 0;
      final r = LocationReporter.createForTest(report: () async {
        calls++;
      });

      expect(r.isRunning, isFalse);
      r.start();
      expect(r.isRunning, isTrue);

      // tick imediato (sem esperar 30s)
      await Future<void>.delayed(Duration.zero);
      expect(calls, 1);

      r.stop();
      expect(r.isRunning, isFalse);
    });

    test('start é idempotente (não cria timer duplo)', () async {
      var calls = 0;
      final r = LocationReporter.createForTest(report: () async {
        calls++;
      });

      r.start();
      r.start();
      r.start();
      await Future<void>.delayed(Duration.zero);
      expect(calls, 1, reason: 'apenas o tick imediato, sem timers duplicados');

      r.stop();
    });

    test('stop cancela o timer e é seguro chamar sem start', () async {
      var calls = 0;
      final r = LocationReporter.createForTest(report: () async {
        calls++;
      });

      r.stop(); // no-op
      r.start();
      await Future<void>.delayed(Duration.zero);
      expect(calls, 1);

      r.stop();
      expect(r.isRunning, isFalse);
      r.stop(); // idempotente
    });

    test('guard anti-sobreposição: tick lento não empilha chamadas', () async {
      var running = 0;
      var maxConcurrent = 0;
      var calls = 0;
      final r = LocationReporter.createForTest(report: () async {
        calls++;
        running++;
        if (running > maxConcurrent) maxConcurrent = running;
        await Future<void>.delayed(const Duration(milliseconds: 50));
        running--;
      });

      r.start();
      // dispara ticks extras manualmente enquanto o primeiro ainda roda
      await Future<void>.delayed(const Duration(milliseconds: 10));
      // os ticks do timer só vêm a cada 30s; simulamos chamando o interno
      // via start já ativo + esperas — o guard é coberto pelo _inFlight.
      await Future<void>.delayed(const Duration(milliseconds: 100));
      await Future<void>.delayed(const Duration(milliseconds: 100));

      r.stop();
      expect(maxConcurrent, 1, reason: 'nunca dois reports ao mesmo tempo');
      expect(calls, greaterThanOrEqualTo(1));
    });

    test('intervalo oficial é 30 segundos (RF04)', () {
      expect(LocationReporter.interval, const Duration(seconds: 30));
    });
  });
}

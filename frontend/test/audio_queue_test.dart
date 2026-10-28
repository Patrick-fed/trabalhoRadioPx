import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/voice/services/audio_queue.dart';

Uint8List bytes(int n) => Uint8List.fromList(List.filled(n, 1));

void main() {
  group('AudioQueue (RF08)', () {
    test('reproduz chunks em ordem FIFO e para ao esvaziar', () async {
      final queue = AudioQueue();
      final received = <Uint8List>[];
      queue.audioStream.listen(received.add);

      queue.enqueue(bytes(3), senderId: 'a');
      queue.enqueue(bytes(2), senderId: 'a');
      queue.enqueue(bytes(1), senderId: 'a');
      queue.startPlayback();

      await Future<void>.delayed(const Duration(milliseconds: 300));

      expect(received.length, 3);
      expect(received[0].length, 3);
      expect(received[1].length, 2);
      expect(received[2].length, 1);
      expect(queue.queueLength, 0);
      expect(queue.isPlaying, isFalse);

      queue.dispose();
    });

    test('teto de memória descarta os chunks mais antigos', () {
      final queue = AudioQueue(maxBytes: 10);

      queue.enqueue(bytes(4));
      queue.enqueue(bytes(4));
      expect(queue.queueLength, 2);
      expect(queue.queuedBytes, 8);

      queue.enqueue(bytes(4)); // 12 > 10: descarta o mais antigo
      expect(queue.queueLength, 2);
      expect(queue.queuedBytes, 8);

      queue.dispose();
    });

    test('sentinel dispara endOfSegment sem emitir áudio', () async {
      final queue = AudioQueue();
      final received = <Uint8List>[];
      var ended = 0;
      queue.audioStream.listen(received.add);
      queue.endOfSegmentStream.listen((_) => ended++);

      queue.enqueue(bytes(4));
      queue.enqueueEndOfSegment();
      queue.enqueue(bytes(2));
      queue.startPlayback();

      await Future<void>.delayed(const Duration(milliseconds: 300));

      expect(received.length, 2);
      expect(received[0].length, 4);
      expect(received[1].length, 2);
      expect(ended, 1);

      queue.dispose();
    });

    test('ignora chunk vazio e mantém o último acima do teto', () {
      final queue = AudioQueue(maxBytes: 5);

      queue.enqueue(Uint8List(0));
      expect(queue.queueLength, 0);

      queue.enqueue(bytes(9)); // único chunk: mantido mesmo acima do teto
      expect(queue.queueLength, 1);
      expect(queue.queuedBytes, 9);

      queue.dispose();
    });

    test('clear esvazia a fila e para a reprodução', () {
      final queue = AudioQueue();
      queue.enqueue(bytes(4));
      queue.startPlayback();
      expect(queue.isPlaying, isTrue);

      queue.clear();
      expect(queue.queueLength, 0);
      expect(queue.queuedBytes, 0);
      expect(queue.isPlaying, isFalse);

      queue.dispose();
    });
  });
}

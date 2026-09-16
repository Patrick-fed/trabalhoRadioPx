import 'package:flutter_test/flutter_test.dart';
import 'package:radiopx/features/auth/models/user_model.dart';
import 'package:radiopx/features/channels/models/channel_model.dart';
import 'package:radiopx/features/voice/models/audio_packet.dart';

void main() {
  group('UserModel', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': '123',
        'email': 'test@example.com',
        'name': 'Test User',
        'avatar': null,
        'provider': 'local',
        'provider_id': null,
        'is_active': true,
        'created_at': '2026-01-01T00:00:00.000Z',
        'updated_at': '2026-01-01T00:00:00.000Z',
      };

      final user = UserModel.fromJson(json);

      expect(user.id, '123');
      expect(user.email, 'test@example.com');
      expect(user.name, 'Test User');
      expect(user.provider, 'local');
      expect(user.isActive, true);
      expect(user.createdAt, isNotNull);
    });

    test('toJson serializes correctly', () {
      const user = UserModel(
        id: '123',
        email: 'test@example.com',
        name: 'Test User',
      );

      final json = user.toJson();

      expect(json['id'], '123');
      expect(json['email'], 'test@example.com');
      expect(json['name'], 'Test User');
    });

    test('displayName returns name when available', () {
      const user = UserModel(
        id: '1',
        email: 'test@example.com',
        name: 'John',
      );
      expect(user.displayName, 'John');
    });

    test('displayName returns email prefix when name is null', () {
      const user = UserModel(id: '1', email: 'john@example.com');
      expect(user.displayName, 'john');
    });

    test('initials from name', () {
      const user = UserModel(
        id: '1',
        email: 'test@example.com',
        name: 'John Doe',
      );
      expect(user.initials, 'J');
    });

    test('initials from email when name is null', () {
      const user = UserModel(id: '1', email: 'alice@example.com');
      expect(user.initials, 'A');
    });

    test('copyWith preserves unchanged fields', () {
      const user = UserModel(
        id: '1',
        email: 'test@example.com',
        name: 'Original',
      );
      final updated = user.copyWith(name: 'Updated');

      expect(updated.id, '1');
      expect(updated.email, 'test@example.com');
      expect(updated.name, 'Updated');
    });

    test('equality based on id', () {
      const user1 = UserModel(id: '1', email: 'a@test.com');
      const user2 = UserModel(id: '1', email: 'b@test.com');
      const user3 = UserModel(id: '2', email: 'a@test.com');

      expect(user1, equals(user2));
      expect(user1, isNot(equals(user3)));
    });
  });

  group('ChannelModel', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'ch-1',
        'name': 'Canal Teste',
        'description': 'Descricao',
        'latitude': -23.55,
        'longitude': -46.63,
        'user_count': 5,
        'max_users': 10,
        'is_available': true,
        'distance': 2.5,
        'created_by': 'user-1',
      };

      final channel = ChannelModel.fromJson(json);

      expect(channel.id, 'ch-1');
      expect(channel.name, 'Canal Teste');
      expect(channel.latitude, -23.55);
      expect(channel.longitude, -46.63);
      expect(channel.userCount, 5);
      expect(channel.maxUsers, 10);
      expect(channel.isAvailable, true);
      expect(channel.distance, 2.5);
    });

    test('toJson serializes correctly', () {
      const channel = ChannelModel(
        id: 'ch-1',
        name: 'Canal',
        latitude: -23.55,
        longitude: -46.63,
      );

      final json = channel.toJson();

      expect(json['id'], 'ch-1');
      expect(json['name'], 'Canal');
      expect(json['latitude'], -23.55);
      expect(json['longitude'], -46.63);
    });

    test('distanceFormatted returns meters for small distances', () {
      const channel = ChannelModel(id: '1', name: 'C', distance: 0.5);
      expect(channel.distanceFormatted, '500m');
    });

    test('distanceFormatted returns km for larger distances', () {
      const channel = ChannelModel(id: '1', name: 'C', distance: 3.7);
      expect(channel.distanceFormatted, '3.7km');
    });

    test('distanceFormatted returns empty when null', () {
      const channel = ChannelModel(id: '1', name: 'C');
      expect(channel.distanceFormatted, '');
    });

    test('equality based on id', () {
      const ch1 = ChannelModel(id: '1', name: 'A');
      const ch2 = ChannelModel(id: '1', name: 'B');
      const ch3 = ChannelModel(id: '2', name: 'A');

      expect(ch1, equals(ch2));
      expect(ch1, isNot(equals(ch3)));
    });
  });

  group('AudioPacket', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'pkt-1',
        'user_id': 'user-1',
        'channel_id': 'ch-1',
        'audio': [1, 2, 3],
        'is_ptt': true,
        'timestamp': '2026-01-01T00:00:00.000Z',
        'codec': 'opus',
      };

      final packet = AudioPacket.fromJson(json);

      expect(packet.id, 'pkt-1');
      expect(packet.userId, 'user-1');
      expect(packet.channelId, 'ch-1');
      expect(packet.audio, [1, 2, 3]);
      expect(packet.isPtt, true);
      expect(packet.codec, 'opus');
    });

    test('toJson serializes correctly', () {
      final packet = AudioPacket(
        id: 'pkt-1',
        userId: 'user-1',
        channelId: 'ch-1',
        audio: [10, 20],
        isPtt: false,
        timestamp: DateTime(2026, 1, 1),
      );

      final json = packet.toJson();

      expect(json['id'], 'pkt-1');
      expect(json['user_id'], 'user-1');
      expect(json['channel_id'], 'ch-1');
      expect(json['audio'], [10, 20]);
      expect(json['is_ptt'], false);
    });
  });

  group('AudioMessage', () {
    test('fromJson parses correctly', () {
      final json = {
        'type': 'audio',
        'payload': {'data': 'base64audio'},
      };

      final msg = AudioMessage.fromJson(json);

      expect(msg.type, 'audio');
      expect(msg.payload['data'], 'base64audio');
    });
  });

  group('VoiceStatus', () {
    test('fromJson parses correctly', () {
      final json = {
        'is_transmitting': true,
        'current_speaker': 'user-1',
        'channel_id': 'ch-1',
      };

      final status = VoiceStatus.fromJson(json);

      expect(status.isTransmitting, true);
      expect(status.currentSpeaker, 'user-1');
      expect(status.channelId, 'ch-1');
    });

    test('toJson serializes correctly', () {
      final status = VoiceStatus(
        isTransmitting: false,
        channelId: 'ch-1',
      );

      final json = status.toJson();

      expect(json['is_transmitting'], false);
      expect(json['current_speaker'], null);
      expect(json['channel_id'], 'ch-1');
    });
  });
}

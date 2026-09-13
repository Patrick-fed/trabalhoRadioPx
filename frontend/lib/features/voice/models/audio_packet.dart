class AudioPacket {
  final String id;
  final String userId;
  final String channelId;
  final List<int> audio;
  final bool isPtt;
  final DateTime timestamp;
  final String codec;

  AudioPacket({
    required this.id,
    required this.userId,
    required this.channelId,
    required this.audio,
    required this.isPtt,
    required this.timestamp,
    this.codec = 'opus',
  });

  factory AudioPacket.fromJson(Map<String, dynamic> json) {
    return AudioPacket(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      channelId: json['channel_id'] as String,
      audio: List<int>.from(json['audio'] as List),
      isPtt: json['is_ptt'] as bool,
      timestamp: DateTime.parse(json['timestamp'] as String),
      codec: json['codec'] as String? ?? 'opus',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'channel_id': channelId,
      'audio': audio,
      'is_ptt': isPtt,
      'timestamp': timestamp.toIso8601String(),
      'codec': codec,
    };
  }
}

class AudioMessage {
  final String type;
  final Map<String, dynamic> payload;

  AudioMessage({
    required this.type,
    required this.payload,
  });

  factory AudioMessage.fromJson(Map<String, dynamic> json) {
    return AudioMessage(
      type: json['type'] as String,
      payload: json['payload'] as Map<String, dynamic>,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'payload': payload,
    };
  }
}

class VoiceStatus {
  final bool isTransmitting;
  final String? currentSpeaker;
  final String channelId;

  VoiceStatus({
    required this.isTransmitting,
    this.currentSpeaker,
    required this.channelId,
  });

  factory VoiceStatus.fromJson(Map<String, dynamic> json) {
    return VoiceStatus(
      isTransmitting: json['is_transmitting'] as bool,
      currentSpeaker: json['current_speaker'] as String?,
      channelId: json['channel_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'is_transmitting': isTransmitting,
      'current_speaker': currentSpeaker,
      'channel_id': channelId,
    };
  }
}

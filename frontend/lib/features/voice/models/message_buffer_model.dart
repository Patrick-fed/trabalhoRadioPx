import 'dart:typed_data';

class MessageBufferModel {
  final String id;
  final String channelId;
  final String userId;
  final Uint8List audioData;
  final int sequence;
  final DateTime timestamp;
  final bool isReplayed;
  final DateTime createdAt;

  const MessageBufferModel({
    required this.id,
    required this.channelId,
    required this.userId,
    required this.audioData,
    required this.sequence,
    required this.timestamp,
    this.isReplayed = false,
    required this.createdAt,
  });

  factory MessageBufferModel.fromJson(Map<String, dynamic> json) {
    return MessageBufferModel(
      id: json['id'] as String,
      channelId: json['channel_id'] as String,
      userId: json['user_id'] as String,
      audioData: Uint8List.fromList(json['audio_data'] as List<int>),
      sequence: json['sequence'] as int,
      timestamp: DateTime.parse(json['timestamp'] as String),
      isReplayed: json['is_replayed'] as bool? ?? false,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'channel_id': channelId,
      'user_id': userId,
      'audio_data': audioData.toList(),
      'sequence': sequence,
      'timestamp': timestamp.toIso8601String(),
      'is_replayed': isReplayed,
      'created_at': createdAt.toIso8601String(),
    };
  }

  MessageBufferModel copyWith({
    String? id,
    String? channelId,
    String? userId,
    Uint8List? audioData,
    int? sequence,
    DateTime? timestamp,
    bool? isReplayed,
    DateTime? createdAt,
  }) {
    return MessageBufferModel(
      id: id ?? this.id,
      channelId: channelId ?? this.channelId,
      userId: userId ?? this.userId,
      audioData: audioData ?? this.audioData,
      sequence: sequence ?? this.sequence,
      timestamp: timestamp ?? this.timestamp,
      isReplayed: isReplayed ?? this.isReplayed,
      createdAt: createdAt ?? this.createdAt,
    );
  }

  @override
  String toString() {
    return 'MessageBufferModel(id: $id, channelId: $channelId, sequence: $sequence)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is MessageBufferModel && other.id == id;
  }

  @override
  int get hashCode => id.hashCode;
}

class BufferConfigModel {
  final Duration maxDuration;
  final int maxSizeBytes;
  final Duration cleanupInterval;

  const BufferConfigModel({
    this.maxDuration = const Duration(seconds: 30),
    this.maxSizeBytes = 10 * 1024 * 1024, // 10MB
    this.cleanupInterval = const Duration(minutes: 5),
  });

  factory BufferConfigModel.fromJson(Map<String, dynamic> json) {
    return BufferConfigModel(
      maxDuration: Duration(seconds: json['max_duration'] as int? ?? 30),
      maxSizeBytes: json['max_size_bytes'] as int? ?? 10 * 1024 * 1024,
      cleanupInterval: Duration(seconds: json['cleanup_interval'] as int? ?? 300),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'max_duration': maxDuration.inSeconds,
      'max_size_bytes': maxSizeBytes,
      'cleanup_interval': cleanupInterval.inSeconds,
    };
  }
}

class BufferStatusModel {
  final int totalMessages;
  final int totalSizeBytes;
  final DateTime? oldestMessage;
  final DateTime? newestMessage;
  final int channelCount;

  const BufferStatusModel({
    required this.totalMessages,
    required this.totalSizeBytes,
    this.oldestMessage,
    this.newestMessage,
    required this.channelCount,
  });

  factory BufferStatusModel.fromJson(Map<String, dynamic> json) {
    return BufferStatusModel(
      totalMessages: json['total_messages'] as int,
      totalSizeBytes: json['total_size_bytes'] as int,
      oldestMessage: json['oldest_message'] != null
          ? DateTime.parse(json['oldest_message'] as String)
          : null,
      newestMessage: json['newest_message'] != null
          ? DateTime.parse(json['newest_message'] as String)
          : null,
      channelCount: json['channel_count'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'total_messages': totalMessages,
      'total_size_bytes': totalSizeBytes,
      'oldest_message': oldestMessage?.toIso8601String(),
      'newest_message': newestMessage?.toIso8601String(),
      'channel_count': channelCount,
    };
  }

  String get totalSizeFormatted {
    if (totalSizeBytes < 1024) {
      return '${totalSizeBytes}B';
    } else if (totalSizeBytes < 1024 * 1024) {
      return '${(totalSizeBytes / 1024).toStringAsFixed(1)}KB';
    } else {
      return '${(totalSizeBytes / (1024 * 1024)).toStringAsFixed(1)}MB';
    }
  }
}

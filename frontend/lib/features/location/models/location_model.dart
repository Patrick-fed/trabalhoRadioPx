class LocationModel {
  final String? id;
  final String? userId;
  final double latitude;
  final double longitude;
  final double? accuracy;
  final DateTime? timestamp;
  final DateTime? updatedAt;

  const LocationModel({
    this.id,
    this.userId,
    required this.latitude,
    required this.longitude,
    this.accuracy,
    this.timestamp,
    this.updatedAt,
  });

  factory LocationModel.fromJson(Map<String, dynamic> json) {
    return LocationModel(
      id: json['id'] as String?,
      userId: json['user_id'] as String?,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      accuracy: (json['accuracy'] as num?)?.toDouble(),
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'latitude': latitude,
      'longitude': longitude,
      'accuracy': accuracy,
      'timestamp': timestamp?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }

  LocationModel copyWith({
    String? id,
    String? userId,
    double? latitude,
    double? longitude,
    double? accuracy,
    DateTime? timestamp,
    DateTime? updatedAt,
  }) {
    return LocationModel(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      accuracy: accuracy ?? this.accuracy,
      timestamp: timestamp ?? this.timestamp,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  String toString() {
    return 'LocationModel(id: $id, userId: $userId, lat: $latitude, lng: $longitude)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is LocationModel &&
        other.id == id &&
        other.userId == userId &&
        other.latitude == latitude &&
        other.longitude == longitude;
  }

  @override
  int get hashCode {
    return id.hashCode ^ userId.hashCode ^ latitude.hashCode ^ longitude.hashCode;
  }
}

class NearbyChannelModel {
  final String channelId;
  final String name;
  final double latitude;
  final double longitude;
  final double distance;
  final int userCount;
  final int maxUsers;
  final bool isAvailable;

  const NearbyChannelModel({
    required this.channelId,
    required this.name,
    required this.latitude,
    required this.longitude,
    required this.distance,
    required this.userCount,
    required this.maxUsers,
    required this.isAvailable,
  });

  factory NearbyChannelModel.fromJson(Map<String, dynamic> json) {
    return NearbyChannelModel(
      channelId: json['channel_id'] as String,
      name: json['name'] as String,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      distance: (json['distance'] as num).toDouble(),
      userCount: json['user_count'] as int,
      maxUsers: json['max_users'] as int,
      isAvailable: json['is_available'] as bool,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'channel_id': channelId,
      'name': name,
      'latitude': latitude,
      'longitude': longitude,
      'distance': distance,
      'user_count': userCount,
      'max_users': maxUsers,
      'is_available': isAvailable,
    };
  }

  String get distanceFormatted {
    if (distance < 1) {
      return '${(distance * 1000).round()}m';
    }
    return '${distance.toStringAsFixed(1)}km';
  }
}

class ChannelModel {
  final String id;
  final String name;
  final String? description;
  final double? latitude;
  final double? longitude;
  final int userCount;
  final int maxUsers;
  final bool isAvailable;
  final double? distance;
  final String? createdBy;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  const ChannelModel({
    required this.id,
    required this.name,
    this.description,
    this.latitude,
    this.longitude,
    this.userCount = 0,
    this.maxUsers = 10,
    this.isAvailable = true,
    this.distance,
    this.createdBy,
    this.createdAt,
    this.updatedAt,
  });

  factory ChannelModel.fromJson(Map<String, dynamic> json) {
    return ChannelModel(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      userCount: json['user_count'] as int? ?? 0,
      maxUsers: json['max_users'] as int? ?? 10,
      isAvailable: json['is_available'] as bool? ?? true,
      distance: (json['distance'] as num?)?.toDouble(),
      createdBy: json['created_by'] as String?,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'description': description,
      'latitude': latitude,
      'longitude': longitude,
      'user_count': userCount,
      'max_users': maxUsers,
      'is_available': isAvailable,
      'distance': distance,
      'created_by': createdBy,
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }

  ChannelModel copyWith({
    String? id,
    String? name,
    String? description,
    double? latitude,
    double? longitude,
    int? userCount,
    int? maxUsers,
    bool? isAvailable,
    double? distance,
    String? createdBy,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ChannelModel(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      userCount: userCount ?? this.userCount,
      maxUsers: maxUsers ?? this.maxUsers,
      isAvailable: isAvailable ?? this.isAvailable,
      distance: distance ?? this.distance,
      createdBy: createdBy ?? this.createdBy,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  String get distanceFormatted {
    if (distance == null) return '';
    if (distance! < 1) {
      return '${(distance! * 1000).round()}m';
    }
    return '${distance!.toStringAsFixed(1)}km';
  }

  @override
  String toString() {
    return 'ChannelModel(id: $id, name: $name, userCount: $userCount)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is ChannelModel && other.id == id;
  }

  @override
  int get hashCode => id.hashCode;
}

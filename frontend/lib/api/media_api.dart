import 'dart:typed_data';

import 'api_client.dart';

class MediaUploadDto {
  final String url;
  final String key;
  final String contentType;
  final int size;

  const MediaUploadDto({
    required this.url,
    required this.key,
    required this.contentType,
    required this.size,
  });

  factory MediaUploadDto.fromJson(Map<String, dynamic> json) => MediaUploadDto(
        url: json['url'] as String? ?? '',
        key: json['key'] as String? ?? '',
        contentType: json['content_type'] as String? ?? '',
        size: json['size'] as int? ?? 0,
      );
}

class MediaApi {
  final ApiClient _client;

  const MediaApi(this._client);

  Future<MediaUploadDto> uploadImage({
    required String filename,
    required Uint8List bytes,
    required String contentType,
  }) async {
    final json = await _client.uploadBytes(
      '/editor/media/images',
      fieldName: 'file',
      filename: filename,
      bytes: bytes,
      contentType: contentType,
      auth: true,
    );
    return MediaUploadDto.fromJson(json as Map<String, dynamic>);
  }
}

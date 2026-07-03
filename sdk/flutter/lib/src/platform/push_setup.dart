import 'package:flutter/services.dart';

/// Platform-specific push notification setup via method channels.
///
/// Android: Requires FCM setup in `android/app/src/main/`.
/// iOS: Requires APNs capability in Xcode + Push Notifications entitlement.
class PushSetup {
  static const _channel = MethodChannel('him_flutter/push');

  /// Request notification permissions (iOS only).
  static Future<bool> requestPermission() async {
    try {
      final result = await _channel.invokeMethod<bool>('requestPermission');
      return result ?? false;
    } on MissingPluginException {
      return false;
    }
  }

  /// Get the FCM token (Android) or APNs token (iOS).
  static Future<String?> getToken() async {
    try {
      return await _channel.invokeMethod<String>('getToken');
    } on MissingPluginException {
      return null;
    }
  }

  /// Set up a listener for push notification tap events.
  static void onMessageTap(void Function(String) callback) {
    _channel.setMethodCallHandler((call) async {
      if (call.method == 'onMessageTap') {
        callback(call.arguments as String? ?? '');
      }
    });
  }
}

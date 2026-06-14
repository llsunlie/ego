import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:hive_flutter/hive_flutter.dart';

class LocalStore {
  static const _authBox = 'auth';
  static const _settingsBox = 'settings';
  static final _secure = FlutterSecureStorage();
  static String? _cachedToken;
  static String? _cachedRefreshToken;
  static late Box _auth;
  static late Box _settings;

  static Future<void> init() async {
    await Hive.initFlutter();
    _auth = await Hive.openBox(_authBox);
    _settings = await Hive.openBox(_settingsBox);
    if (kIsWeb) {
      // Web: Hive (dart2js incompatible with flutter_secure_storage_web)
      _cachedToken = _auth.get('access_token');
      _cachedRefreshToken = _auth.get('refresh_token');
    } else {
      // Native: Keystore/EncryptedSharedPreferences
      _cachedToken = await _secure.read(key: 'access_token');
      _cachedRefreshToken = await _secure.read(key: 'refresh_token');
    }
  }

  // Auth — native → secure storage, web → Hive
  static String? getToken() => _cachedToken;

  static Future<void> setToken(String token) async {
    if (kIsWeb) {
      _auth.put('access_token', token);
    } else {
      await _secure.write(key: 'access_token', value: token);
    }
    _cachedToken = token;
  }

  static Future<void> clearToken() async {
    if (kIsWeb) {
      _auth.delete('access_token');
      _auth.delete('refresh_token');
    } else {
      await _secure.delete(key: 'access_token');
      await _secure.delete(key: 'refresh_token');
    }
    _cachedToken = null;
    _cachedRefreshToken = null;
  }

  // Refresh token
  static String? getRefreshToken() => _cachedRefreshToken;

  static Future<void> setRefreshToken(String token) async {
    if (kIsWeb) {
      _auth.put('refresh_token', token);
    } else {
      await _secure.write(key: 'refresh_token', value: token);
    }
    _cachedRefreshToken = token;
  }

  static Future<void> clearRefreshToken() async {
    if (kIsWeb) {
      _auth.delete('refresh_token');
    } else {
      await _secure.delete(key: 'refresh_token');
    }
    _cachedRefreshToken = null;
  }

  // Settings
  static bool getOnboardingDone() =>
      _settings.get('onboardingDone', defaultValue: false);
  static void setOnboardingDone(bool value) =>
      _settings.put('onboardingDone', value);

  static bool getStarmapTapGuideShown() =>
      _settings.get('starmapTapGuideShown', defaultValue: false);
  static void setStarmapTapGuideShown(bool value) =>
      _settings.put('starmapTapGuideShown', value);
}

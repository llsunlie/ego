import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:hive_flutter/hive_flutter.dart';

class LocalStore {
  static const _authBox = 'auth';
  static const _settingsBox = 'settings';
  static final _secure = FlutterSecureStorage();
  static String? _cachedToken;
  static late Box _auth;
  static late Box _settings;

  static Future<void> init() async {
    await Hive.initFlutter();
    _auth = await Hive.openBox(_authBox);
    _settings = await Hive.openBox(_settingsBox);
    if (kIsWeb) {
      // Web: Hive (dart2js incompatible with flutter_secure_storage_web)
      _cachedToken = _auth.get('token');
    } else {
      // Native: Keystore/EncryptedSharedPreferences
      _cachedToken = await _secure.read(key: 'token');
    }
  }

  // Auth — native → secure storage, web → Hive
  static String? getToken() => _cachedToken;

  static Future<void> setToken(String token) async {
    if (kIsWeb) {
      _auth.put('token', token);
    } else {
      await _secure.write(key: 'token', value: token);
    }
    _cachedToken = token;
  }

  static Future<void> clearToken() async {
    if (kIsWeb) {
      _auth.delete('token');
    } else {
      await _secure.delete(key: 'token');
    }
    _cachedToken = null;
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

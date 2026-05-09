import 'package:hive_flutter/hive_flutter.dart';

class LocalStore {
  static const _authBox = 'auth';
  static const _settingsBox = 'settings';
  static late Box _auth;
  static late Box _settings;

  static Future<void> init() async {
    await Hive.initFlutter();
    _auth = await Hive.openBox(_authBox);
    _settings = await Hive.openBox(_settingsBox);
  }

  // Auth
  static String? getToken() => _auth.get('token');
  static void setToken(String token) => _auth.put('token', token);
  static void clearToken() => _auth.delete('token');

  // Settings
  static bool getOnboardingDone() =>
      _settings.get('onboardingDone', defaultValue: false);
  static void setOnboardingDone(bool value) =>
      _settings.put('onboardingDone', value);
}

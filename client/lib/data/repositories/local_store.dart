import 'package:hive_flutter/hive_flutter.dart';

class LocalStore {
  static const _boxName = 'auth';
  static late Box _box;

  static Future<void> init() async {
    await Hive.initFlutter();
    _box = await Hive.openBox(_boxName);
  }

  static String? getToken() => _box.get('token');
  static void setToken(String token) => _box.put('token', token);
  static void clearToken() => _box.delete('token');
}

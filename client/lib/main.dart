import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'app.dart';
import 'data/repositories/local_store.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await LocalStore.init();

  runApp(
    const ProviderScope(
      child: App(),
    ),
  );
}

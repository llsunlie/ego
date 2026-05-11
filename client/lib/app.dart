import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:device_preview/device_preview.dart';
import 'core/router/router.dart';
import 'core/theme/theme.dart';
import 'shared/widgets/toast_overlay.dart';

class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return DevicePreview(
      enabled: !kReleaseMode,
      builder: (context) => MaterialApp.router(
        title: 'ego',
        theme: darkTheme(),
        routerConfig: router,
        debugShowCheckedModeBanner: false,
        // ignore: deprecated_member_use
        useInheritedMediaQuery: true,
        locale: DevicePreview.locale(context),
        builder: (context, child) => DevicePreview.appBuilder(
          context,
          ToastOverlay(child: child!),
        ),
      ),
    );
  }
}

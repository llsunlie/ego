import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../features/login/login_page.dart';
import '../../features/now/now_page.dart';
import '../../features/past/past_page.dart';
import '../../features/starmap/starmap_page.dart';
import '../../shared/widgets/app_shell.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/now',
    redirect: (context, state) {
      final loggedIn = authState.isLoggedIn;
      final isLoginPage = state.matchedLocation == '/login';

      if (!loggedIn && !isLoginPage) return '/login';
      if (loggedIn && isLoginPage) return '/now';
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginPage(),
      ),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) =>
            AppShell(navigationShell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/now',
                builder: (context, state) => const NowPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/past',
                builder: (context, state) => const PastPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/starmap',
                builder: (context, state) => const StarmapPage(),
              ),
            ],
          ),
        ],
      ),
    ],
  );
});

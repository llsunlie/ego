import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/providers/onboarding_provider.dart';
import '../../features/login/login_page.dart';
import '../../features/onboarding/onboarding_page.dart';
import '../../features/now/now_page.dart';
import '../../features/past/past_page.dart';
import '../../features/past/trace_detail_page.dart';
import '../../features/starmap/starmap_page.dart';
import '../../features/starmap/constellation_detail_page.dart';
import '../../shared/widgets/app_shell.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);
  final onboardingDone = ref.watch(onboardingCompleteProvider);

  return GoRouter(
    initialLocation:
        Uri.base.fragment.isNotEmpty ? Uri.base.fragment : '/now',
    redirect: (context, state) {
      final loggedIn = authState.isLoggedIn;
      final isLoginRoute = state.matchedLocation == '/login';
      final isOnboardingRoute = state.matchedLocation == '/onboard';

      if (!loggedIn && !isLoginRoute) return '/login';
      if (loggedIn && isLoginRoute) {
        return onboardingDone ? '/now' : '/onboard';
      }
      if (loggedIn && !onboardingDone && !isOnboardingRoute) return '/onboard';
      if (loggedIn && onboardingDone && isOnboardingRoute) return '/now';
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginPage(),
      ),
      GoRoute(
        path: '/onboard',
        builder: (context, state) => const OnboardingPage(),
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
                routes: [
                  GoRoute(
                    path: 'detail/:traceId',
                    builder: (context, state) => TraceDetailPage(
                      traceId: state.pathParameters['traceId']!,
                    ),
                  ),
                ],
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/starmap',
                builder: (context, state) => const StarmapPage(),
                routes: [
                  GoRoute(
                    path: 'detail/:constellationId',
                    builder: (context, state) => ConstellationDetailPage(
                      constellationId: state.pathParameters['constellationId']!,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ],
      ),
    ],
  );
});

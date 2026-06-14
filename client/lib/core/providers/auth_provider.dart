import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/repositories/local_store.dart';

class AuthState {
  final String? accessToken;
  final String? refreshToken;
  final bool isLoggedIn;

  const AuthState({this.accessToken, this.refreshToken, this.isLoggedIn = false});

  AuthState copyWith({String? accessToken, String? refreshToken, bool? isLoggedIn}) {
    return AuthState(
      accessToken: accessToken ?? this.accessToken,
      refreshToken: refreshToken ?? this.refreshToken,
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier() : super(const AuthState()) {
    _loadToken();
  }

  void _loadToken() {
    final token = LocalStore.getToken();
    final refreshToken = LocalStore.getRefreshToken();
    if (token != null) {
      state = AuthState(accessToken: token, refreshToken: refreshToken, isLoggedIn: true);
    }
  }

  Future<void> login(String accessToken, String refreshToken) async {
    await LocalStore.setToken(accessToken);
    await LocalStore.setRefreshToken(refreshToken);
    state = AuthState(accessToken: accessToken, refreshToken: refreshToken, isLoggedIn: true);
  }

  void refreshAccessToken(String accessToken) {
    LocalStore.setToken(accessToken); // fire-and-forget
    state = state.copyWith(accessToken: accessToken);
  }

  Future<void> logout() async {
    await LocalStore.clearToken();
    await LocalStore.clearRefreshToken();
    state = const AuthState();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier();
});

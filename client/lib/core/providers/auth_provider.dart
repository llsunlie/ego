import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/repositories/local_store.dart';

class AuthState {
  final String? token;
  final bool isLoggedIn;

  const AuthState({this.token, this.isLoggedIn = false});

  AuthState copyWith({String? token, bool? isLoggedIn}) {
    return AuthState(
      token: token ?? this.token,
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
    if (token != null) {
      state = AuthState(token: token, isLoggedIn: true);
    }
  }

  void login(String token) {
    LocalStore.setToken(token);
    state = AuthState(token: token, isLoggedIn: true);
  }

  void logout() {
    LocalStore.clearToken();
    state = const AuthState();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier();
});

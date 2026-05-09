import 'package:flutter_riverpod/flutter_riverpod.dart';

class ToastState {
  final String? message;
  final bool isVisible;

  const ToastState({this.message, this.isVisible = false});
}

class ToastNotifier extends StateNotifier<ToastState> {
  ToastNotifier() : super(const ToastState());

  void show(String message) {
    state = ToastState(message: message, isVisible: true);
  }

  void hide() {
    state = const ToastState();
  }
}

final toastProvider = StateNotifierProvider<ToastNotifier, ToastState>((ref) {
  return ToastNotifier();
});

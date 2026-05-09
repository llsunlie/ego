import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/repositories/local_store.dart';

class OnboardingCompleteNotifier extends StateNotifier<bool> {
  OnboardingCompleteNotifier() : super(LocalStore.getOnboardingDone());

  void complete() {
    LocalStore.setOnboardingDone(true);
    state = true;
  }
}

final onboardingCompleteProvider =
    StateNotifierProvider<OnboardingCompleteNotifier, bool>((ref) {
  return OnboardingCompleteNotifier();
});

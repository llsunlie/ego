import 'package:flutter_riverpod/flutter_riverpod.dart';

class TabNotifier extends StateNotifier<int> {
  TabNotifier() : super(0);
  void setIndex(int index) => state = index;
}

final tabProvider = StateNotifierProvider<TabNotifier, int>((ref) {
  return TabNotifier();
});

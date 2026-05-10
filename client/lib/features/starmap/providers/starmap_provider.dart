import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../data/generated/api.pb.dart' as pb;
import '../../../data/services/ego_client.dart';

class StarmapState {
  final List<pb.Constellation> constellations;
  final int totalStarCount;
  final bool isLoading;
  final String? error;

  const StarmapState({
    this.constellations = const [],
    this.totalStarCount = 0,
    this.isLoading = false,
    this.error,
  });

  StarmapState copyWith({
    List<pb.Constellation>? constellations,
    int? totalStarCount,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) {
    return StarmapState(
      constellations: constellations ?? this.constellations,
      totalStarCount: totalStarCount ?? this.totalStarCount,
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : error ?? this.error,
    );
  }
}

class StarmapNotifier extends StateNotifier<StarmapState> {
  final Ref _ref;

  StarmapNotifier(this._ref) : super(const StarmapState());

  EgoClient get _client => _ref.read(EgoClient.provider);

  Future<void> loadConstellations() async {
    if (state.isLoading) return;
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final res = await _client.listConstellations(_ref);
      state = state.copyWith(
        constellations: res.constellations,
        totalStarCount: res.totalStarCount,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> refresh() => loadConstellations();
}

final starmapProvider = StateNotifierProvider<StarmapNotifier, StarmapState>(
  (ref) => StarmapNotifier(ref),
);

final pendingTopicPromptProvider = StateProvider<String?>((ref) => null);

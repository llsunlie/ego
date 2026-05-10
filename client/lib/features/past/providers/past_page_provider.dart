import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../data/generated/api.pb.dart' as pb;
import '../../../data/services/ego_client.dart';

class PastPageState {
  final List<pb.Trace> traces;
  final String? nextCursor;
  final bool hasMore;
  final bool isLoading;
  final bool isLoadingMore;
  final String? error;

  const PastPageState({
    this.traces = const [],
    this.nextCursor,
    this.hasMore = false,
    this.isLoading = false,
    this.isLoadingMore = false,
    this.error,
  });

  PastPageState copyWith({
    List<pb.Trace>? traces,
    String? nextCursor,
    bool? hasMore,
    bool? isLoading,
    bool? isLoadingMore,
    String? error,
    bool clearError = false,
  }) {
    return PastPageState(
      traces: traces ?? this.traces,
      nextCursor: nextCursor ?? this.nextCursor,
      hasMore: hasMore ?? this.hasMore,
      isLoading: isLoading ?? this.isLoading,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      error: clearError ? null : error ?? this.error,
    );
  }

  Map<String, List<pb.Trace>> get monthGroups {
    final groups = <String, List<pb.Trace>>{};
    for (final t in traces) {
      final dt = DateTime.fromMillisecondsSinceEpoch(t.createdAt.toInt());
      final key = '${dt.year}年${dt.month}月';
      groups.putIfAbsent(key, () => []).add(t);
    }
    return groups;
  }

  List<String> get sortedMonthKeys {
    final keys = monthGroups.keys.toList();
    keys.sort((a, b) => b.compareTo(a));
    return keys;
  }
}

class PastPageNotifier extends StateNotifier<PastPageState> {
  final Ref _ref;

  PastPageNotifier(this._ref) : super(const PastPageState());

  EgoClient get _client => _ref.read(EgoClient.provider);

  Future<void> loadFirstPage() async {
    if (state.isLoading) return;
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final res = await _client.listTraces(
        _ref,
        cursor: '',
        pageSize: 20,
      );
      state = state.copyWith(
        traces: res.traces,
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> loadNextPage() async {
    if (state.isLoadingMore || !state.hasMore) return;
    state = state.copyWith(isLoadingMore: true);
    try {
      final res = await _client.listTraces(
        _ref,
        cursor: state.nextCursor ?? '',
        pageSize: 20,
      );
      state = state.copyWith(
        traces: [...state.traces, ...res.traces],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        isLoadingMore: false,
      );
    } catch (e) {
      state = state.copyWith(isLoadingMore: false, error: e.toString());
    }
  }

  Future<void> refresh() async {
    await loadFirstPage();
  }
}

final pastPageProvider =
    StateNotifierProvider<PastPageNotifier, PastPageState>(
  (ref) => PastPageNotifier(ref),
);

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../data/generated/api.pb.dart' as pb;
import '../../../data/services/ego_client.dart';

enum NowPageStatus { idle, writing, echoing, stashing }

class NowPageState {
  final NowPageStatus status;
  final String? currentTraceId;
  final String? currentMomentId;
  final pb.Echo? echo;
  final pb.Insight? insight;
  final bool isLoading;
  final String? error;
  final bool isReopen;
  final List<pb.Moment> matchedMoments;

  const NowPageState({
    this.status = NowPageStatus.idle,
    this.currentTraceId,
    this.currentMomentId,
    this.echo,
    this.insight,
    this.isLoading = false,
    this.error,
    this.isReopen = false,
    this.matchedMoments = const [],
  });

  NowPageState copyWith({
    NowPageStatus? status,
    String? currentTraceId,
    String? currentMomentId,
    pb.Echo? echo,
    pb.Insight? insight,
    bool? isLoading,
    String? error,
    bool? isReopen,
    List<pb.Moment>? matchedMoments,
    bool clearTraceId = false,
    bool clearMomentId = false,
    bool clearEcho = false,
    bool clearInsight = false,
    bool clearError = false,
    bool clearIsReopen = false,
    bool clearMatchedMoments = false,
  }) {
    return NowPageState(
      status: status ?? this.status,
      currentTraceId:
          clearTraceId ? null : currentTraceId ?? this.currentTraceId,
      currentMomentId:
          clearMomentId ? null : currentMomentId ?? this.currentMomentId,
      echo: clearEcho ? null : echo ?? this.echo,
      insight: clearInsight ? null : insight ?? this.insight,
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : error ?? this.error,
      isReopen:
          clearIsReopen ? false : isReopen ?? this.isReopen,
      matchedMoments:
          clearMatchedMoments ? const [] : matchedMoments ?? this.matchedMoments,
    );
  }
}

class NowPageNotifier extends StateNotifier<NowPageState> {
  final Ref _ref;

  NowPageNotifier(this._ref) : super(const NowPageState());

  EgoClient get _client => _ref.read(EgoClient.provider);

  void startWriting() {
    state = state.copyWith(status: NowPageStatus.writing, clearIsReopen: true);
  }

  void reopenWhisper() {
    state = state.copyWith(status: NowPageStatus.writing, isReopen: true);
  }

  Future<void> submitMoment(String content) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final res = await _client.createMoment(
        _ref,
        content: content,
        traceId: state.currentTraceId,
      );
      final echo = res.hasEcho() ? res.echo : null;
      state = state.copyWith(
        status: NowPageStatus.echoing,
        currentTraceId: res.moment.traceId,
        currentMomentId: res.moment.id,
        echo: echo,
        isLoading: false,
        clearIsReopen: true,
      );

      _fetchInsight(res.moment.id, echo?.id ?? '');
      if (echo != null && echo.matchedMomentIds.isNotEmpty) {
        _fetchMatchedMoments(echo.matchedMomentIds);
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  Future<void> _fetchInsight(String momentId, String echoId) async {
    try {
      final res = await _client.generateInsight(
        _ref,
        momentId: momentId,
        echoId: echoId,
      );
      if (res.hasInsight()) {
        state = state.copyWith(insight: res.insight);
      }
    } catch (_) {
      // Insight is optional — silent failure
    }
  }

  Future<void> _fetchMatchedMoments(List<String> ids) async {
    try {
      final res = await _client.getMoments(_ref, ids: ids);
      final moments = res.moments.toList();
      // Reorder to match input IDs order (GetMoments may return arbitrary order)
      final byId = <String, pb.Moment>{};
      for (final m in moments) {
        byId[m.id] = m;
      }
      final ordered = <pb.Moment>[];
      for (final id in ids) {
        if (byId.containsKey(id)) ordered.add(byId[id]!);
      }
      state = state.copyWith(matchedMoments: ordered);
    } catch (_) {
      // Matched moments fetch is optional — silent failure
    }
  }

  void beginStash() {
    if (state.currentTraceId == null) return;
    state = state.copyWith(status: NowPageStatus.stashing);
  }

  Future<void> completeStash() async {
    final traceId = state.currentTraceId;
    if (traceId != null) {
      try {
        await _client.stashTrace(_ref, traceId: traceId);
      } catch (_) {
        // Stash failure shouldn't block reset
      }
    }
    _resetToIdle();
  }

  void dismissEcho() => _resetToIdle();

  void _resetToIdle() {
    state = const NowPageState();
    _ref.invalidate(memoryDotsProvider);
  }
}

final nowPageProvider =
    StateNotifierProvider<NowPageNotifier, NowPageState>((ref) {
  return NowPageNotifier(ref);
});

final memoryDotsProvider = FutureProvider.autoDispose<List<pb.Moment>>((ref) async {
  final client = ref.read(EgoClient.provider);
  final res = await client.getRandomMoments(ref, count: 3);
  return res.moments.toList();
});

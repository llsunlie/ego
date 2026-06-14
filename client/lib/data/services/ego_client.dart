import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc_or_grpcweb.dart';
import '../generated/api.pbgrpc.dart' as grpc;
import 'interceptors/auth_interceptor.dart';
import '../repositories/local_store.dart';
import '../../core/constants.dart';
import '../../core/providers/auth_provider.dart';

class EgoClient {
  final grpc.EgoClient _stub;

  EgoClient(this._stub);

  grpc.EgoClient get stub => _stub;

  static final provider = Provider<EgoClient>((ref) {
    final channel = GrpcOrGrpcWebClientChannel.toSeparatePorts(
      host: AppConstants.serverHost,
      grpcPort: AppConstants.serverPort,
      grpcTransportSecure: AppConstants.serverTls,
      grpcWebPort: AppConstants.serverWebPort,
      grpcWebTransportSecure: AppConstants.serverTls,
    );
    return EgoClient(grpc.EgoClient(channel));
  });

  CallOptions _withAuth(dynamic ref) {
    final token = ref.read(authProvider).accessToken;
    return authCallOptions(token);
  }

  /// Wraps an authenticated gRPC call with automatic token refresh on 401.
  /// On UNAUTHENTICATED: attempts RefreshToken → retries once → logs out on failure.
  Future<T> _autoRefresh<T>(dynamic ref, Future<T> Function() grpcCall) async {
    try {
      return await grpcCall();
    } on GrpcError catch (e) {
      if (e.code != StatusCode.unauthenticated) rethrow;

      final storedRefresh = LocalStore.getRefreshToken();
      if (storedRefresh == null) rethrow;

      try {
        final res = await refreshToken(storedRefresh);
        ref.read(authProvider.notifier).refreshAccessToken(res.accessToken);
        return grpcCall(); // Retry once with new token
      } catch (_) {
        ref.read(authProvider.notifier).logout();
        rethrow;
      }
    }
  }

  // ─── RefreshToken ──────────────────────────────

  Future<grpc.RefreshTokenRes> refreshToken(String token) async {
    final req = grpc.RefreshTokenReq(refreshToken: token);
    return _stub.refreshToken(req);
  }

  // ─── Auth ─────────────────────────────────────

  Future<grpc.CheckPhoneRes> checkPhone(String phone) async {
    final req = grpc.CheckPhoneReq(phone: phone);
    return _stub.checkPhone(req);
  }

  Future<grpc.SendVerificationCodeRes> sendVerificationCode(String phone) async {
    final req = grpc.SendVerificationCodeReq(phone: phone);
    return _stub.sendVerificationCode(req);
  }

  Future<grpc.RegisterRes> register({
    required String phone,
    required String code,
    required String password,
  }) async {
    final req = grpc.RegisterReq(phone: phone, code: code, password: password);
    return _stub.register(req);
  }

  Future<grpc.LoginRes> login(String phone, String password) async {
    final req = grpc.LoginReq(phone: phone, password: password);
    return _stub.login(req);
  }

  Future<grpc.ResetPasswordRes> resetPassword({
    required String phone,
    required String code,
    required String newPassword,
  }) async {
    final req = grpc.ResetPasswordReq(phone: phone, code: code, newPassword: newPassword);
    return _stub.resetPassword(req);
  }

  // ─── Setting ───────────────────────────────────

  Future<grpc.GetProfileRes> getProfile(dynamic ref) async {
    return _autoRefresh(ref, () {
      final req = grpc.GetProfileReq();
      return _stub.getProfile(req, options: _withAuth(ref));
    });
  }

  Future<grpc.SubmitFeedbackRes> submitFeedback(
    dynamic ref, {
    required String content,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.SubmitFeedbackReq(content: content);
      return _stub.submitFeedback(req, options: _withAuth(ref));
    });
  }

  // ─── Moment ───────────────────────────────────

  Future<grpc.CreateMomentRes> createMoment(
    dynamic ref, {
    required String content,
    String? traceId,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.CreateMomentReq(content: content, traceId: traceId ?? '');
      return _stub.createMoment(req, options: _withAuth(ref));
    });
  }

  Future<grpc.GetMomentsRes> getMoments(
    dynamic ref, {
    required List<String> ids,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.GetMomentsReq(ids: ids);
      return _stub.getMoments(req, options: _withAuth(ref));
    });
  }

  Future<grpc.GenerateInsightRes> generateInsight(
    dynamic ref, {
    required String momentId,
    required String echoId,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.GenerateInsightReq(momentId: momentId, echoId: echoId);
      return _stub.generateInsight(req, options: _withAuth(ref));
    });
  }

  // ─── Trace ────────────────────────────────────

  Future<grpc.ListTracesRes> listTraces(
    dynamic ref, {
    String cursor = '',
    int pageSize = 20,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.ListTracesReq(cursor: cursor, pageSize: pageSize);
      return _stub.listTraces(req, options: _withAuth(ref));
    });
  }

  Future<grpc.GetTraceDetailRes> getTraceDetail(
    dynamic ref, {
    required String traceId,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.GetTraceDetailReq(traceId: traceId);
      return _stub.getTraceDetail(req, options: _withAuth(ref));
    });
  }

  // ─── Memory Dot ───────────────────────────────

  Future<grpc.GetRandomMomentsRes> getRandomMoments(
    dynamic ref, {
    int count = 3,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.GetRandomMomentsReq(count: count);
      return _stub.getRandomMoments(req, options: _withAuth(ref));
    });
  }

  // ─── Stash ────────────────────────────────────

  Future<grpc.StashTraceRes> stashTrace(
    dynamic ref, {
    required String traceId,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.StashTraceReq(traceId: traceId);
      return _stub.stashTrace(req, options: _withAuth(ref));
    });
  }

  // ─── Constellation ────────────────────────────

  Future<grpc.ListConstellationsRes> listConstellations(dynamic ref) async {
    return _autoRefresh(ref, () {
      final req = grpc.ListConstellationsReq();
      return _stub.listConstellations(req, options: _withAuth(ref));
    });
  }

  Future<grpc.GetConstellationRes> getConstellation(
    dynamic ref, {
    required String constellationId,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.GetConstellationReq(constellationId: constellationId);
      return _stub.getConstellation(req, options: _withAuth(ref));
    });
  }

  // ─── Chat ─────────────────────────────────────

  Future<grpc.StartChatRes> startChat(
    dynamic ref, {
    required String starId,
    String chatSessionId = '',
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.StartChatReq(
        starId: starId,
        chatSessionId: chatSessionId,
      );
      return _stub.startChat(req, options: _withAuth(ref));
    });
  }

  Future<grpc.SendMessageRes> sendMessage(
    dynamic ref, {
    required String chatSessionId,
    required String content,
  }) async {
    return _autoRefresh(ref, () {
      final req = grpc.SendMessageReq(
        chatSessionId: chatSessionId,
        content: content,
      );
      return _stub.sendMessage(req, options: _withAuth(ref));
    });
  }
}

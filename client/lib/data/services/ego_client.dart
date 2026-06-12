import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc_or_grpcweb.dart';
import '../generated/api.pbgrpc.dart' as grpc;
import 'interceptors/auth_interceptor.dart';
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

  CallOptions _withAuth(Ref ref) {
    final token = ref.read(authProvider).token;
    return authCallOptions(token);
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

  Future<grpc.GetProfileRes> getProfile(WidgetRef ref) async {
    final req = grpc.GetProfileReq();
    final token = ref.read(authProvider).token;
    return _stub.getProfile(req, options: authCallOptions(token));
  }

  Future<grpc.SubmitFeedbackRes> submitFeedback(
    WidgetRef ref, {
    required String content,
  }) async {
    final req = grpc.SubmitFeedbackReq(content: content);
    final token = ref.read(authProvider).token;
    return _stub.submitFeedback(req, options: authCallOptions(token));
  }

  // ─── Moment ───────────────────────────────────

  Future<grpc.CreateMomentRes> createMoment(
    Ref ref, {
    required String content,
    String? traceId,
  }) async {
    final req = grpc.CreateMomentReq(content: content, traceId: traceId ?? '');
    return _stub.createMoment(req, options: _withAuth(ref));
  }

  Future<grpc.GetMomentsRes> getMoments(
    Ref ref, {
    required List<String> ids,
  }) async {
    final req = grpc.GetMomentsReq(ids: ids);
    return _stub.getMoments(req, options: _withAuth(ref));
  }

  Future<grpc.GenerateInsightRes> generateInsight(
    Ref ref, {
    required String momentId,
    required String echoId,
  }) async {
    final req = grpc.GenerateInsightReq(momentId: momentId, echoId: echoId);
    return _stub.generateInsight(req, options: _withAuth(ref));
  }

  // ─── Trace ────────────────────────────────────

  Future<grpc.ListTracesRes> listTraces(
    Ref ref, {
    String cursor = '',
    int pageSize = 20,
  }) async {
    final req = grpc.ListTracesReq(cursor: cursor, pageSize: pageSize);
    return _stub.listTraces(req, options: _withAuth(ref));
  }

  Future<grpc.GetTraceDetailRes> getTraceDetail(
    Ref ref, {
    required String traceId,
  }) async {
    final req = grpc.GetTraceDetailReq(traceId: traceId);
    return _stub.getTraceDetail(req, options: _withAuth(ref));
  }

  // ─── Memory Dot ───────────────────────────────

  Future<grpc.GetRandomMomentsRes> getRandomMoments(
    Ref ref, {
    int count = 3,
  }) async {
    final req = grpc.GetRandomMomentsReq(count: count);
    return _stub.getRandomMoments(req, options: _withAuth(ref));
  }

  // ─── Stash ────────────────────────────────────

  Future<grpc.StashTraceRes> stashTrace(
    Ref ref, {
    required String traceId,
  }) async {
    final req = grpc.StashTraceReq(traceId: traceId);
    return _stub.stashTrace(req, options: _withAuth(ref));
  }

  // ─── Constellation ────────────────────────────

  Future<grpc.ListConstellationsRes> listConstellations(Ref ref) async {
    final req = grpc.ListConstellationsReq();
    return _stub.listConstellations(req, options: _withAuth(ref));
  }

  Future<grpc.GetConstellationRes> getConstellation(
    Ref ref, {
    required String constellationId,
  }) async {
    final req = grpc.GetConstellationReq(constellationId: constellationId);
    return _stub.getConstellation(req, options: _withAuth(ref));
  }

  // ─── Chat ─────────────────────────────────────

  Future<grpc.StartChatRes> startChat(
    Ref ref, {
    required String starId,
    String chatSessionId = '',
  }) async {
    final req = grpc.StartChatReq(
      starId: starId,
      chatSessionId: chatSessionId,
    );
    return _stub.startChat(req, options: _withAuth(ref));
  }

  Future<grpc.SendMessageRes> sendMessage(
    Ref ref, {
    required String chatSessionId,
    required String content,
  }) async {
    final req = grpc.SendMessageReq(
      chatSessionId: chatSessionId,
      content: content,
    );
    return _stub.sendMessage(req, options: _withAuth(ref));
  }
}

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc_or_grpcweb.dart';
import '../generated/api.pbgrpc.dart' as grpc;
import 'interceptors/auth_interceptor.dart';
import '../../core/constants.dart';
import '../../core/providers/auth_provider.dart';

class EgoClient {
  final grpc.EgoClient _stub;

  EgoClient(this._stub);

  static final provider = Provider<EgoClient>((ref) {
    final channel = GrpcOrGrpcWebClientChannel.toSeparatePorts(
      host: AppConstants.serverHost,
      grpcPort: AppConstants.serverPort,
      grpcTransportSecure: false,
      grpcWebPort: AppConstants.serverWebPort,
      grpcWebTransportSecure: false,
    );
    return EgoClient(grpc.EgoClient(channel));
  });

  // Wired for Module 2+; used when calling authenticated RPCs.
  // ignore: unused_element
  CallOptions _withAuth(Ref ref) {
    final token = ref.read(authProvider).token;
    return authCallOptions(token);
  }

  Future<grpc.LoginRes> login(String account, String password) async {
    final req = grpc.LoginReq(account: account, password: password);
    return _stub.login(req);
  }
}

import 'package:grpc/grpc.dart';

/// Inject Authorization: Bearer <token> metadata into gRPC calls
CallOptions authCallOptions(String? token) {
  if (token == null) return CallOptions();
  return CallOptions(
    metadata: {'Authorization': 'Bearer $token'},
  );
}

import 'package:grpc/grpc_or_grpcweb.dart';

// Re-export GrpcError so that files importing this mapper can use
// `on GrpcError catch (e)` without an additional direct import.
export 'package:grpc/grpc_or_grpcweb.dart' show GrpcError;

/// Maps gRPC errors to user-facing Chinese messages.
///
/// Prefers the server's own message (e.message) — the server is the authority
/// on user-facing error text and returns Chinese messages for all known error
/// conditions. Falls back to code-based defaults only when the server provides
/// no message.
String grpcErrorMessage(GrpcError e) {
  // Use the server's message if it provides a meaningful one.
  if (e.message != null && e.message!.isNotEmpty) {
    return e.message!;
  }

  // Fallback: code-based defaults when server provides no message.
  switch (e.code) {
    case StatusCode.unauthenticated:
      return '认证失败，请检查登录状态';
    case StatusCode.unavailable:
      return '服务暂不可用，请稍后重试';
    case StatusCode.deadlineExceeded:
      return '请求超时，请检查网络后重试';
    case StatusCode.notFound:
      return '请求的资源不存在';
    case StatusCode.alreadyExists:
      return '资源已存在';
    case StatusCode.invalidArgument:
      return '请求参数有误';
    case StatusCode.internal:
      return '服务器内部错误，请稍后重试';
    case StatusCode.permissionDenied:
      return '权限不足';
    default:
      return '网络错误，请稍后重试';
  }
}

/// Extracts a user-friendly message from any exception.
///
/// If the exception is a [GrpcError], uses [grpcErrorMessage].
/// Otherwise returns [fallback].
String errorMessage(Object e, {String fallback = '网络错误，请稍后重试'}) {
  if (e is GrpcError) return grpcErrorMessage(e);
  return fallback;
}

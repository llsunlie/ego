// This is a generated file - do not edit.
//
// Generated from ego/api.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'api.pb.dart' as $0;

export 'api.pb.dart';

@$pb.GrpcServiceName('ego.Ego')
class EgoClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  EgoClient(super.channel, {super.options, super.interceptors});

  $grpc.ResponseFuture<$0.LoginRes> login(
    $0.LoginReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$login, request, options: options);
  }

  // method descriptors

  static final _$login = $grpc.ClientMethod<$0.LoginReq, $0.LoginRes>(
      '/ego.Ego/Login',
      ($0.LoginReq value) => value.writeToBuffer(),
      $0.LoginRes.fromBuffer);
}

@$pb.GrpcServiceName('ego.Ego')
abstract class EgoServiceBase extends $grpc.Service {
  $core.String get $name => 'ego.Ego';

  EgoServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.LoginReq, $0.LoginRes>(
        'Login',
        login_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LoginReq.fromBuffer(value),
        ($0.LoginRes value) => value.writeToBuffer()));
  }

  $async.Future<$0.LoginRes> login_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.LoginReq> $request) async {
    return login($call, await $request);
  }

  $async.Future<$0.LoginRes> login($grpc.ServiceCall call, $0.LoginReq request);
}

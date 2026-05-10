// This is a generated file - do not edit.
//
// Generated from api.proto.

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

  /// ─── Auth（认证）─────────────────────────────────────
  $grpc.ResponseFuture<$0.LoginRes> login(
    $0.LoginReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$login, request, options: options);
  }

  /// ─── Moment（话语）─────────────────────────────────────
  /// 写下一句话，保存 Moment + Echo，返回回声
  $grpc.ResponseFuture<$0.CreateMomentRes> createMoment(
    $0.CreateMomentReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createMoment, request, options: options);
  }

  /// 根据 ID 列表批量获取 Moment 内容
  $grpc.ResponseFuture<$0.GetMomentsRes> getMoments(
    $0.GetMomentsReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getMoments, request, options: options);
  }

  /// 获取 AI 观察（per-Moment，结合 Echo）
  $grpc.ResponseFuture<$0.GenerateInsightRes> generateInsight(
    $0.GenerateInsightReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$generateInsight, request, options: options);
  }

  /// ─── Trace（写作会话）─────────────────────────────────
  /// 获取过往 Trace 列表（游标分页，按月分组）
  $grpc.ResponseFuture<$0.ListTracesRes> listTraces(
    $0.ListTracesReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listTraces, request, options: options);
  }

  /// 获取 Trace 详情（Item[] = <Moment, Echo[], Insight>）
  $grpc.ResponseFuture<$0.GetTraceDetailRes> getTraceDetail(
    $0.GetTraceDetailReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getTraceDetail, request, options: options);
  }

  /// ─── Memory Dot（记忆光点盲盒）────────────────────────
  /// 随机获取 N 条历史话语
  $grpc.ResponseFuture<$0.GetRandomMomentsRes> getRandomMoments(
    $0.GetRandomMomentsReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getRandomMoments, request, options: options);
  }

  /// ─── Stash（收进星图）─────────────────────────────────
  /// 将 Trace 寄存为 Star，触发异步聚类
  $grpc.ResponseFuture<$0.StashTraceRes> stashTrace(
    $0.StashTraceReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$stashTrace, request, options: options);
  }

  /// ─── Constellation（星座）──────────────────────────────
  /// 获取所有星座列表（星图页渲染用）
  $grpc.ResponseFuture<$0.ListConstellationsRes> listConstellations(
    $0.ListConstellationsReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listConstellations, request, options: options);
  }

  /// 获取星座详情（含 insight、Stars、topic_prompts）
  $grpc.ResponseFuture<$0.GetConstellationRes> getConstellation(
    $0.GetConstellationReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getConstellation, request, options: options);
  }

  /// ─── Chat（和那时的自己说说话）──────────────────────────
  /// 开始一次对话（后端构建 past-self 上下文）
  $grpc.ResponseFuture<$0.StartChatRes> startChat(
    $0.StartChatReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$startChat, request, options: options);
  }

  /// 发送消息并获取回复
  $grpc.ResponseFuture<$0.SendMessageRes> sendMessage(
    $0.SendMessageReq request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$sendMessage, request, options: options);
  }

  // method descriptors

  static final _$login = $grpc.ClientMethod<$0.LoginReq, $0.LoginRes>(
      '/ego.Ego/Login',
      ($0.LoginReq value) => value.writeToBuffer(),
      $0.LoginRes.fromBuffer);
  static final _$createMoment =
      $grpc.ClientMethod<$0.CreateMomentReq, $0.CreateMomentRes>(
          '/ego.Ego/CreateMoment',
          ($0.CreateMomentReq value) => value.writeToBuffer(),
          $0.CreateMomentRes.fromBuffer);
  static final _$getMoments =
      $grpc.ClientMethod<$0.GetMomentsReq, $0.GetMomentsRes>(
          '/ego.Ego/GetMoments',
          ($0.GetMomentsReq value) => value.writeToBuffer(),
          $0.GetMomentsRes.fromBuffer);
  static final _$generateInsight =
      $grpc.ClientMethod<$0.GenerateInsightReq, $0.GenerateInsightRes>(
          '/ego.Ego/GenerateInsight',
          ($0.GenerateInsightReq value) => value.writeToBuffer(),
          $0.GenerateInsightRes.fromBuffer);
  static final _$listTraces =
      $grpc.ClientMethod<$0.ListTracesReq, $0.ListTracesRes>(
          '/ego.Ego/ListTraces',
          ($0.ListTracesReq value) => value.writeToBuffer(),
          $0.ListTracesRes.fromBuffer);
  static final _$getTraceDetail =
      $grpc.ClientMethod<$0.GetTraceDetailReq, $0.GetTraceDetailRes>(
          '/ego.Ego/GetTraceDetail',
          ($0.GetTraceDetailReq value) => value.writeToBuffer(),
          $0.GetTraceDetailRes.fromBuffer);
  static final _$getRandomMoments =
      $grpc.ClientMethod<$0.GetRandomMomentsReq, $0.GetRandomMomentsRes>(
          '/ego.Ego/GetRandomMoments',
          ($0.GetRandomMomentsReq value) => value.writeToBuffer(),
          $0.GetRandomMomentsRes.fromBuffer);
  static final _$stashTrace =
      $grpc.ClientMethod<$0.StashTraceReq, $0.StashTraceRes>(
          '/ego.Ego/StashTrace',
          ($0.StashTraceReq value) => value.writeToBuffer(),
          $0.StashTraceRes.fromBuffer);
  static final _$listConstellations =
      $grpc.ClientMethod<$0.ListConstellationsReq, $0.ListConstellationsRes>(
          '/ego.Ego/ListConstellations',
          ($0.ListConstellationsReq value) => value.writeToBuffer(),
          $0.ListConstellationsRes.fromBuffer);
  static final _$getConstellation =
      $grpc.ClientMethod<$0.GetConstellationReq, $0.GetConstellationRes>(
          '/ego.Ego/GetConstellation',
          ($0.GetConstellationReq value) => value.writeToBuffer(),
          $0.GetConstellationRes.fromBuffer);
  static final _$startChat =
      $grpc.ClientMethod<$0.StartChatReq, $0.StartChatRes>(
          '/ego.Ego/StartChat',
          ($0.StartChatReq value) => value.writeToBuffer(),
          $0.StartChatRes.fromBuffer);
  static final _$sendMessage =
      $grpc.ClientMethod<$0.SendMessageReq, $0.SendMessageRes>(
          '/ego.Ego/SendMessage',
          ($0.SendMessageReq value) => value.writeToBuffer(),
          $0.SendMessageRes.fromBuffer);
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
    $addMethod($grpc.ServiceMethod<$0.CreateMomentReq, $0.CreateMomentRes>(
        'CreateMoment',
        createMoment_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CreateMomentReq.fromBuffer(value),
        ($0.CreateMomentRes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetMomentsReq, $0.GetMomentsRes>(
        'GetMoments',
        getMoments_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.GetMomentsReq.fromBuffer(value),
        ($0.GetMomentsRes value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.GenerateInsightReq, $0.GenerateInsightRes>(
            'GenerateInsight',
            generateInsight_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.GenerateInsightReq.fromBuffer(value),
            ($0.GenerateInsightRes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ListTracesReq, $0.ListTracesRes>(
        'ListTraces',
        listTraces_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ListTracesReq.fromBuffer(value),
        ($0.ListTracesRes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetTraceDetailReq, $0.GetTraceDetailRes>(
        'GetTraceDetail',
        getTraceDetail_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.GetTraceDetailReq.fromBuffer(value),
        ($0.GetTraceDetailRes value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.GetRandomMomentsReq, $0.GetRandomMomentsRes>(
            'GetRandomMoments',
            getRandomMoments_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.GetRandomMomentsReq.fromBuffer(value),
            ($0.GetRandomMomentsRes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.StashTraceReq, $0.StashTraceRes>(
        'StashTrace',
        stashTrace_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.StashTraceReq.fromBuffer(value),
        ($0.StashTraceRes value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.ListConstellationsReq, $0.ListConstellationsRes>(
            'ListConstellations',
            listConstellations_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ListConstellationsReq.fromBuffer(value),
            ($0.ListConstellationsRes value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.GetConstellationReq, $0.GetConstellationRes>(
            'GetConstellation',
            getConstellation_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.GetConstellationReq.fromBuffer(value),
            ($0.GetConstellationRes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.StartChatReq, $0.StartChatRes>(
        'StartChat',
        startChat_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.StartChatReq.fromBuffer(value),
        ($0.StartChatRes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SendMessageReq, $0.SendMessageRes>(
        'SendMessage',
        sendMessage_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.SendMessageReq.fromBuffer(value),
        ($0.SendMessageRes value) => value.writeToBuffer()));
  }

  $async.Future<$0.LoginRes> login_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.LoginReq> $request) async {
    return login($call, await $request);
  }

  $async.Future<$0.LoginRes> login($grpc.ServiceCall call, $0.LoginReq request);

  $async.Future<$0.CreateMomentRes> createMoment_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateMomentReq> $request) async {
    return createMoment($call, await $request);
  }

  $async.Future<$0.CreateMomentRes> createMoment(
      $grpc.ServiceCall call, $0.CreateMomentReq request);

  $async.Future<$0.GetMomentsRes> getMoments_Pre($grpc.ServiceCall $call,
      $async.Future<$0.GetMomentsReq> $request) async {
    return getMoments($call, await $request);
  }

  $async.Future<$0.GetMomentsRes> getMoments(
      $grpc.ServiceCall call, $0.GetMomentsReq request);

  $async.Future<$0.GenerateInsightRes> generateInsight_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GenerateInsightReq> $request) async {
    return generateInsight($call, await $request);
  }

  $async.Future<$0.GenerateInsightRes> generateInsight(
      $grpc.ServiceCall call, $0.GenerateInsightReq request);

  $async.Future<$0.ListTracesRes> listTraces_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.ListTracesReq> $request) async {
    return listTraces($call, await $request);
  }

  $async.Future<$0.ListTracesRes> listTraces(
      $grpc.ServiceCall call, $0.ListTracesReq request);

  $async.Future<$0.GetTraceDetailRes> getTraceDetail_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetTraceDetailReq> $request) async {
    return getTraceDetail($call, await $request);
  }

  $async.Future<$0.GetTraceDetailRes> getTraceDetail(
      $grpc.ServiceCall call, $0.GetTraceDetailReq request);

  $async.Future<$0.GetRandomMomentsRes> getRandomMoments_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetRandomMomentsReq> $request) async {
    return getRandomMoments($call, await $request);
  }

  $async.Future<$0.GetRandomMomentsRes> getRandomMoments(
      $grpc.ServiceCall call, $0.GetRandomMomentsReq request);

  $async.Future<$0.StashTraceRes> stashTrace_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.StashTraceReq> $request) async {
    return stashTrace($call, await $request);
  }

  $async.Future<$0.StashTraceRes> stashTrace(
      $grpc.ServiceCall call, $0.StashTraceReq request);

  $async.Future<$0.ListConstellationsRes> listConstellations_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.ListConstellationsReq> $request) async {
    return listConstellations($call, await $request);
  }

  $async.Future<$0.ListConstellationsRes> listConstellations(
      $grpc.ServiceCall call, $0.ListConstellationsReq request);

  $async.Future<$0.GetConstellationRes> getConstellation_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetConstellationReq> $request) async {
    return getConstellation($call, await $request);
  }

  $async.Future<$0.GetConstellationRes> getConstellation(
      $grpc.ServiceCall call, $0.GetConstellationReq request);

  $async.Future<$0.StartChatRes> startChat_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.StartChatReq> $request) async {
    return startChat($call, await $request);
  }

  $async.Future<$0.StartChatRes> startChat(
      $grpc.ServiceCall call, $0.StartChatReq request);

  $async.Future<$0.SendMessageRes> sendMessage_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SendMessageReq> $request) async {
    return sendMessage($call, await $request);
  }

  $async.Future<$0.SendMessageRes> sendMessage(
      $grpc.ServiceCall call, $0.SendMessageReq request);
}

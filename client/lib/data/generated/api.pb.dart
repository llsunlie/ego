// This is a generated file - do not edit.
//
// Generated from ego/api.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'api.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'api.pbenum.dart';

class LoginReq extends $pb.GeneratedMessage {
  factory LoginReq({
    $core.String? account,
    $core.String? password,
  }) {
    final result = create();
    if (account != null) result.account = account;
    if (password != null) result.password = password;
    return result;
  }

  LoginReq._();

  factory LoginReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'account')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginReq copyWith(void Function(LoginReq) updates) =>
      super.copyWith((message) => updates(message as LoginReq)) as LoginReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginReq create() => LoginReq._();
  @$core.override
  LoginReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static LoginReq getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginReq>(create);
  static LoginReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get account => $_getSZ(0);
  @$pb.TagNumber(1)
  set account($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccount() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

class LoginRes extends $pb.GeneratedMessage {
  factory LoginRes({
    $core.String? token,
    $core.bool? created,
  }) {
    final result = create();
    if (token != null) result.token = token;
    if (created != null) result.created = created;
    return result;
  }

  LoginRes._();

  factory LoginRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..aOB(2, _omitFieldNames ? '' : 'created')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRes copyWith(void Function(LoginRes) updates) =>
      super.copyWith((message) => updates(message as LoginRes)) as LoginRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginRes create() => LoginRes._();
  @$core.override
  LoginRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static LoginRes getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginRes>(create);
  static LoginRes? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get created => $_getBF(1);
  @$pb.TagNumber(2)
  set created($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCreated() => $_has(1);
  @$pb.TagNumber(2)
  void clearCreated() => $_clearField(2);
}

class Moment extends $pb.GeneratedMessage {
  factory Moment({
    $core.String? id,
    $core.String? content,
    $core.String? traceId,
    $fixnum.Int64? createdAt,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (content != null) result.content = content;
    if (traceId != null) result.traceId = traceId;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  Moment._();

  factory Moment.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Moment.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Moment',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'content')
    ..aOS(3, _omitFieldNames ? '' : 'traceId')
    ..aInt64(4, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Moment clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Moment copyWith(void Function(Moment) updates) =>
      super.copyWith((message) => updates(message as Moment)) as Moment;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Moment create() => Moment._();
  @$core.override
  Moment createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Moment getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Moment>(create);
  static Moment? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get content => $_getSZ(1);
  @$pb.TagNumber(2)
  set content($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearContent() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get traceId => $_getSZ(2);
  @$pb.TagNumber(3)
  set traceId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTraceId() => $_has(2);
  @$pb.TagNumber(3)
  void clearTraceId() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get createdAt => $_getI64(3);
  @$pb.TagNumber(4)
  set createdAt($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCreatedAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearCreatedAt() => $_clearField(4);
}

class Echo extends $pb.GeneratedMessage {
  factory Echo({
    $core.String? id,
    $core.String? momentId,
    $core.Iterable<$core.String>? matchedMomentIds,
    $core.Iterable<$core.double>? similarities,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (momentId != null) result.momentId = momentId;
    if (matchedMomentIds != null)
      result.matchedMomentIds.addAll(matchedMomentIds);
    if (similarities != null) result.similarities.addAll(similarities);
    return result;
  }

  Echo._();

  factory Echo.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Echo.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Echo',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'momentId')
    ..pPS(3, _omitFieldNames ? '' : 'matchedMomentIds')
    ..p<$core.double>(
        4, _omitFieldNames ? '' : 'similarities', $pb.PbFieldType.KF)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Echo clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Echo copyWith(void Function(Echo) updates) =>
      super.copyWith((message) => updates(message as Echo)) as Echo;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Echo create() => Echo._();
  @$core.override
  Echo createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Echo getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Echo>(create);
  static Echo? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get momentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set momentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMomentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearMomentId() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get matchedMomentIds => $_getList(2);

  @$pb.TagNumber(4)
  $pb.PbList<$core.double> get similarities => $_getList(3);
}

class Insight extends $pb.GeneratedMessage {
  factory Insight({
    $core.String? id,
    $core.String? momentId,
    $core.String? echoId,
    $core.String? text,
    $core.Iterable<$core.String>? relatedMomentIds,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (momentId != null) result.momentId = momentId;
    if (echoId != null) result.echoId = echoId;
    if (text != null) result.text = text;
    if (relatedMomentIds != null)
      result.relatedMomentIds.addAll(relatedMomentIds);
    return result;
  }

  Insight._();

  factory Insight.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Insight.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Insight',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'momentId')
    ..aOS(3, _omitFieldNames ? '' : 'echoId')
    ..aOS(4, _omitFieldNames ? '' : 'text')
    ..pPS(5, _omitFieldNames ? '' : 'relatedMomentIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Insight clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Insight copyWith(void Function(Insight) updates) =>
      super.copyWith((message) => updates(message as Insight)) as Insight;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Insight create() => Insight._();
  @$core.override
  Insight createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Insight getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Insight>(create);
  static Insight? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get momentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set momentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMomentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearMomentId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get echoId => $_getSZ(2);
  @$pb.TagNumber(3)
  set echoId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEchoId() => $_has(2);
  @$pb.TagNumber(3)
  void clearEchoId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get text => $_getSZ(3);
  @$pb.TagNumber(4)
  set text($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasText() => $_has(3);
  @$pb.TagNumber(4)
  void clearText() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<$core.String> get relatedMomentIds => $_getList(4);
}

class Star extends $pb.GeneratedMessage {
  factory Star({
    $core.String? id,
    $core.String? traceId,
    $core.String? topic,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (traceId != null) result.traceId = traceId;
    if (topic != null) result.topic = topic;
    return result;
  }

  Star._();

  factory Star.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Star.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Star',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'traceId')
    ..aOS(3, _omitFieldNames ? '' : 'topic')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Star clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Star copyWith(void Function(Star) updates) =>
      super.copyWith((message) => updates(message as Star)) as Star;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Star create() => Star._();
  @$core.override
  Star createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Star getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Star>(create);
  static Star? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get traceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set traceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTraceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearTraceId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get topic => $_getSZ(2);
  @$pb.TagNumber(3)
  set topic($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTopic() => $_has(2);
  @$pb.TagNumber(3)
  void clearTopic() => $_clearField(3);
}

class Constellation extends $pb.GeneratedMessage {
  factory Constellation({
    $core.String? id,
    $core.String? name,
    $core.String? constellationInsight,
    $core.Iterable<$core.String>? starIds,
    $core.Iterable<$core.String>? topicPrompts,
    $core.int? starCount,
    $fixnum.Int64? createdAt,
    $fixnum.Int64? updatedAt,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (constellationInsight != null)
      result.constellationInsight = constellationInsight;
    if (starIds != null) result.starIds.addAll(starIds);
    if (topicPrompts != null) result.topicPrompts.addAll(topicPrompts);
    if (starCount != null) result.starCount = starCount;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  Constellation._();

  factory Constellation.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Constellation.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Constellation',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'constellationInsight')
    ..pPS(4, _omitFieldNames ? '' : 'starIds')
    ..pPS(5, _omitFieldNames ? '' : 'topicPrompts')
    ..aI(6, _omitFieldNames ? '' : 'starCount')
    ..aInt64(7, _omitFieldNames ? '' : 'createdAt')
    ..aInt64(8, _omitFieldNames ? '' : 'updatedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Constellation clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Constellation copyWith(void Function(Constellation) updates) =>
      super.copyWith((message) => updates(message as Constellation))
          as Constellation;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Constellation create() => Constellation._();
  @$core.override
  Constellation createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Constellation getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Constellation>(create);
  static Constellation? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get constellationInsight => $_getSZ(2);
  @$pb.TagNumber(3)
  set constellationInsight($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasConstellationInsight() => $_has(2);
  @$pb.TagNumber(3)
  void clearConstellationInsight() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get starIds => $_getList(3);

  @$pb.TagNumber(5)
  $pb.PbList<$core.String> get topicPrompts => $_getList(4);

  @$pb.TagNumber(6)
  $core.int get starCount => $_getIZ(5);
  @$pb.TagNumber(6)
  set starCount($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStarCount() => $_has(5);
  @$pb.TagNumber(6)
  void clearStarCount() => $_clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get createdAt => $_getI64(6);
  @$pb.TagNumber(7)
  set createdAt($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCreatedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearCreatedAt() => $_clearField(7);

  @$pb.TagNumber(8)
  $fixnum.Int64 get updatedAt => $_getI64(7);
  @$pb.TagNumber(8)
  set updatedAt($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasUpdatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearUpdatedAt() => $_clearField(8);
}

class ChatMessage extends $pb.GeneratedMessage {
  factory ChatMessage({
    $core.String? id,
    ChatRole? role,
    $core.String? content,
    $core.Iterable<MomentReference>? referenced,
    $fixnum.Int64? timestamp,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (role != null) result.role = role;
    if (content != null) result.content = content;
    if (referenced != null) result.referenced.addAll(referenced);
    if (timestamp != null) result.timestamp = timestamp;
    return result;
  }

  ChatMessage._();

  factory ChatMessage.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ChatMessage.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChatMessage',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aE<ChatRole>(2, _omitFieldNames ? '' : 'role',
        enumValues: ChatRole.values)
    ..aOS(3, _omitFieldNames ? '' : 'content')
    ..pPM<MomentReference>(4, _omitFieldNames ? '' : 'referenced',
        subBuilder: MomentReference.create)
    ..aInt64(5, _omitFieldNames ? '' : 'timestamp')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatMessage clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatMessage copyWith(void Function(ChatMessage) updates) =>
      super.copyWith((message) => updates(message as ChatMessage))
          as ChatMessage;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ChatMessage create() => ChatMessage._();
  @$core.override
  ChatMessage createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ChatMessage getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChatMessage>(create);
  static ChatMessage? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  ChatRole get role => $_getN(1);
  @$pb.TagNumber(2)
  set role(ChatRole value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRole() => $_has(1);
  @$pb.TagNumber(2)
  void clearRole() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get content => $_getSZ(2);
  @$pb.TagNumber(3)
  set content($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContent() => $_has(2);
  @$pb.TagNumber(3)
  void clearContent() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<MomentReference> get referenced => $_getList(3);

  @$pb.TagNumber(5)
  $fixnum.Int64 get timestamp => $_getI64(4);
  @$pb.TagNumber(5)
  set timestamp($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTimestamp() => $_has(4);
  @$pb.TagNumber(5)
  void clearTimestamp() => $_clearField(5);
}

class MomentReference extends $pb.GeneratedMessage {
  factory MomentReference({
    $core.String? date,
    $core.String? snippet,
  }) {
    final result = create();
    if (date != null) result.date = date;
    if (snippet != null) result.snippet = snippet;
    return result;
  }

  MomentReference._();

  factory MomentReference.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MomentReference.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MomentReference',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'date')
    ..aOS(2, _omitFieldNames ? '' : 'snippet')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MomentReference clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MomentReference copyWith(void Function(MomentReference) updates) =>
      super.copyWith((message) => updates(message as MomentReference))
          as MomentReference;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MomentReference create() => MomentReference._();
  @$core.override
  MomentReference createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static MomentReference getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MomentReference>(create);
  static MomentReference? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get date => $_getSZ(0);
  @$pb.TagNumber(1)
  set date($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDate() => $_has(0);
  @$pb.TagNumber(1)
  void clearDate() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get snippet => $_getSZ(1);
  @$pb.TagNumber(2)
  set snippet($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSnippet() => $_has(1);
  @$pb.TagNumber(2)
  void clearSnippet() => $_clearField(2);
}

class CreateMomentReq extends $pb.GeneratedMessage {
  factory CreateMomentReq({
    $core.String? content,
    $core.String? traceId,
  }) {
    final result = create();
    if (content != null) result.content = content;
    if (traceId != null) result.traceId = traceId;
    return result;
  }

  CreateMomentReq._();

  factory CreateMomentReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateMomentReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateMomentReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'content')
    ..aOS(2, _omitFieldNames ? '' : 'traceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMomentReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMomentReq copyWith(void Function(CreateMomentReq) updates) =>
      super.copyWith((message) => updates(message as CreateMomentReq))
          as CreateMomentReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMomentReq create() => CreateMomentReq._();
  @$core.override
  CreateMomentReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateMomentReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateMomentReq>(create);
  static CreateMomentReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get content => $_getSZ(0);
  @$pb.TagNumber(1)
  set content($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasContent() => $_has(0);
  @$pb.TagNumber(1)
  void clearContent() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get traceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set traceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTraceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearTraceId() => $_clearField(2);
}

class CreateMomentRes extends $pb.GeneratedMessage {
  factory CreateMomentRes({
    Moment? moment,
    Echo? echo,
  }) {
    final result = create();
    if (moment != null) result.moment = moment;
    if (echo != null) result.echo = echo;
    return result;
  }

  CreateMomentRes._();

  factory CreateMomentRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateMomentRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateMomentRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<Moment>(1, _omitFieldNames ? '' : 'moment', subBuilder: Moment.create)
    ..aOM<Echo>(2, _omitFieldNames ? '' : 'echo', subBuilder: Echo.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMomentRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMomentRes copyWith(void Function(CreateMomentRes) updates) =>
      super.copyWith((message) => updates(message as CreateMomentRes))
          as CreateMomentRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMomentRes create() => CreateMomentRes._();
  @$core.override
  CreateMomentRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateMomentRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateMomentRes>(create);
  static CreateMomentRes? _defaultInstance;

  @$pb.TagNumber(1)
  Moment get moment => $_getN(0);
  @$pb.TagNumber(1)
  set moment(Moment value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMoment() => $_has(0);
  @$pb.TagNumber(1)
  void clearMoment() => $_clearField(1);
  @$pb.TagNumber(1)
  Moment ensureMoment() => $_ensure(0);

  @$pb.TagNumber(2)
  Echo get echo => $_getN(1);
  @$pb.TagNumber(2)
  set echo(Echo value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasEcho() => $_has(1);
  @$pb.TagNumber(2)
  void clearEcho() => $_clearField(2);
  @$pb.TagNumber(2)
  Echo ensureEcho() => $_ensure(1);
}

class GetMomentsReq extends $pb.GeneratedMessage {
  factory GetMomentsReq({
    $core.Iterable<$core.String>? ids,
  }) {
    final result = create();
    if (ids != null) result.ids.addAll(ids);
    return result;
  }

  GetMomentsReq._();

  factory GetMomentsReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetMomentsReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMomentsReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'ids')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMomentsReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMomentsReq copyWith(void Function(GetMomentsReq) updates) =>
      super.copyWith((message) => updates(message as GetMomentsReq))
          as GetMomentsReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetMomentsReq create() => GetMomentsReq._();
  @$core.override
  GetMomentsReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetMomentsReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMomentsReq>(create);
  static GetMomentsReq? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get ids => $_getList(0);
}

class GetMomentsRes extends $pb.GeneratedMessage {
  factory GetMomentsRes({
    $core.Iterable<Moment>? moments,
  }) {
    final result = create();
    if (moments != null) result.moments.addAll(moments);
    return result;
  }

  GetMomentsRes._();

  factory GetMomentsRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetMomentsRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMomentsRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..pPM<Moment>(1, _omitFieldNames ? '' : 'moments',
        subBuilder: Moment.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMomentsRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMomentsRes copyWith(void Function(GetMomentsRes) updates) =>
      super.copyWith((message) => updates(message as GetMomentsRes))
          as GetMomentsRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetMomentsRes create() => GetMomentsRes._();
  @$core.override
  GetMomentsRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetMomentsRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMomentsRes>(create);
  static GetMomentsRes? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Moment> get moments => $_getList(0);
}

class GenerateInsightReq extends $pb.GeneratedMessage {
  factory GenerateInsightReq({
    $core.String? momentId,
    $core.String? echoId,
  }) {
    final result = create();
    if (momentId != null) result.momentId = momentId;
    if (echoId != null) result.echoId = echoId;
    return result;
  }

  GenerateInsightReq._();

  factory GenerateInsightReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GenerateInsightReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GenerateInsightReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'momentId')
    ..aOS(2, _omitFieldNames ? '' : 'echoId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateInsightReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateInsightReq copyWith(void Function(GenerateInsightReq) updates) =>
      super.copyWith((message) => updates(message as GenerateInsightReq))
          as GenerateInsightReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GenerateInsightReq create() => GenerateInsightReq._();
  @$core.override
  GenerateInsightReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GenerateInsightReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GenerateInsightReq>(create);
  static GenerateInsightReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get momentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set momentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMomentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMomentId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get echoId => $_getSZ(1);
  @$pb.TagNumber(2)
  set echoId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEchoId() => $_has(1);
  @$pb.TagNumber(2)
  void clearEchoId() => $_clearField(2);
}

class GenerateInsightRes extends $pb.GeneratedMessage {
  factory GenerateInsightRes({
    Insight? insight,
  }) {
    final result = create();
    if (insight != null) result.insight = insight;
    return result;
  }

  GenerateInsightRes._();

  factory GenerateInsightRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GenerateInsightRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GenerateInsightRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<Insight>(1, _omitFieldNames ? '' : 'insight',
        subBuilder: Insight.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateInsightRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateInsightRes copyWith(void Function(GenerateInsightRes) updates) =>
      super.copyWith((message) => updates(message as GenerateInsightRes))
          as GenerateInsightRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GenerateInsightRes create() => GenerateInsightRes._();
  @$core.override
  GenerateInsightRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GenerateInsightRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GenerateInsightRes>(create);
  static GenerateInsightRes? _defaultInstance;

  @$pb.TagNumber(1)
  Insight get insight => $_getN(0);
  @$pb.TagNumber(1)
  set insight(Insight value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasInsight() => $_has(0);
  @$pb.TagNumber(1)
  void clearInsight() => $_clearField(1);
  @$pb.TagNumber(1)
  Insight ensureInsight() => $_ensure(0);
}

class Trace extends $pb.GeneratedMessage {
  factory Trace({
    $core.String? id,
    $core.String? motivation,
    $core.bool? stashed,
    $fixnum.Int64? createdAt,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (motivation != null) result.motivation = motivation;
    if (stashed != null) result.stashed = stashed;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  Trace._();

  factory Trace.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Trace.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Trace',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'motivation')
    ..aOB(3, _omitFieldNames ? '' : 'stashed')
    ..aInt64(4, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Trace clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Trace copyWith(void Function(Trace) updates) =>
      super.copyWith((message) => updates(message as Trace)) as Trace;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Trace create() => Trace._();
  @$core.override
  Trace createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Trace getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Trace>(create);
  static Trace? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get motivation => $_getSZ(1);
  @$pb.TagNumber(2)
  set motivation($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMotivation() => $_has(1);
  @$pb.TagNumber(2)
  void clearMotivation() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get stashed => $_getBF(2);
  @$pb.TagNumber(3)
  set stashed($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStashed() => $_has(2);
  @$pb.TagNumber(3)
  void clearStashed() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get createdAt => $_getI64(3);
  @$pb.TagNumber(4)
  set createdAt($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCreatedAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearCreatedAt() => $_clearField(4);
}

class TraceItem extends $pb.GeneratedMessage {
  factory TraceItem({
    Moment? moment,
    $core.Iterable<Echo>? echos,
    Insight? insight,
  }) {
    final result = create();
    if (moment != null) result.moment = moment;
    if (echos != null) result.echos.addAll(echos);
    if (insight != null) result.insight = insight;
    return result;
  }

  TraceItem._();

  factory TraceItem.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TraceItem.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TraceItem',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<Moment>(1, _omitFieldNames ? '' : 'moment', subBuilder: Moment.create)
    ..pPM<Echo>(2, _omitFieldNames ? '' : 'echos', subBuilder: Echo.create)
    ..aOM<Insight>(3, _omitFieldNames ? '' : 'insight',
        subBuilder: Insight.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TraceItem clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TraceItem copyWith(void Function(TraceItem) updates) =>
      super.copyWith((message) => updates(message as TraceItem)) as TraceItem;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TraceItem create() => TraceItem._();
  @$core.override
  TraceItem createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static TraceItem getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TraceItem>(create);
  static TraceItem? _defaultInstance;

  @$pb.TagNumber(1)
  Moment get moment => $_getN(0);
  @$pb.TagNumber(1)
  set moment(Moment value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMoment() => $_has(0);
  @$pb.TagNumber(1)
  void clearMoment() => $_clearField(1);
  @$pb.TagNumber(1)
  Moment ensureMoment() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<Echo> get echos => $_getList(1);

  @$pb.TagNumber(3)
  Insight get insight => $_getN(2);
  @$pb.TagNumber(3)
  set insight(Insight value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasInsight() => $_has(2);
  @$pb.TagNumber(3)
  void clearInsight() => $_clearField(3);
  @$pb.TagNumber(3)
  Insight ensureInsight() => $_ensure(2);
}

class ListTracesReq extends $pb.GeneratedMessage {
  factory ListTracesReq({
    $core.String? cursor,
    $core.int? pageSize,
  }) {
    final result = create();
    if (cursor != null) result.cursor = cursor;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListTracesReq._();

  factory ListTracesReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListTracesReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTracesReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'cursor')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTracesReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTracesReq copyWith(void Function(ListTracesReq) updates) =>
      super.copyWith((message) => updates(message as ListTracesReq))
          as ListTracesReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListTracesReq create() => ListTracesReq._();
  @$core.override
  ListTracesReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListTracesReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTracesReq>(create);
  static ListTracesReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get cursor => $_getSZ(0);
  @$pb.TagNumber(1)
  set cursor($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCursor() => $_has(0);
  @$pb.TagNumber(1)
  void clearCursor() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get pageSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set pageSize($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageSize() => $_clearField(2);
}

class ListTracesRes extends $pb.GeneratedMessage {
  factory ListTracesRes({
    $core.Iterable<Trace>? traces,
    $core.String? nextCursor,
    $core.bool? hasMore,
  }) {
    final result = create();
    if (traces != null) result.traces.addAll(traces);
    if (nextCursor != null) result.nextCursor = nextCursor;
    if (hasMore != null) result.hasMore = hasMore;
    return result;
  }

  ListTracesRes._();

  factory ListTracesRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListTracesRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTracesRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..pPM<Trace>(1, _omitFieldNames ? '' : 'traces', subBuilder: Trace.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextCursor')
    ..aOB(3, _omitFieldNames ? '' : 'hasMore')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTracesRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTracesRes copyWith(void Function(ListTracesRes) updates) =>
      super.copyWith((message) => updates(message as ListTracesRes))
          as ListTracesRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListTracesRes create() => ListTracesRes._();
  @$core.override
  ListTracesRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListTracesRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTracesRes>(create);
  static ListTracesRes? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Trace> get traces => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextCursor => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextCursor($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextCursor() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextCursor() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get hasMore => $_getBF(2);
  @$pb.TagNumber(3)
  set hasMore($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasHasMore() => $_has(2);
  @$pb.TagNumber(3)
  void clearHasMore() => $_clearField(3);
}

class GetTraceDetailReq extends $pb.GeneratedMessage {
  factory GetTraceDetailReq({
    $core.String? traceId,
  }) {
    final result = create();
    if (traceId != null) result.traceId = traceId;
    return result;
  }

  GetTraceDetailReq._();

  factory GetTraceDetailReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetTraceDetailReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTraceDetailReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'traceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTraceDetailReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTraceDetailReq copyWith(void Function(GetTraceDetailReq) updates) =>
      super.copyWith((message) => updates(message as GetTraceDetailReq))
          as GetTraceDetailReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTraceDetailReq create() => GetTraceDetailReq._();
  @$core.override
  GetTraceDetailReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetTraceDetailReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTraceDetailReq>(create);
  static GetTraceDetailReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get traceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set traceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTraceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTraceId() => $_clearField(1);
}

class GetTraceDetailRes extends $pb.GeneratedMessage {
  factory GetTraceDetailRes({
    Trace? trace,
    $core.Iterable<TraceItem>? items,
  }) {
    final result = create();
    if (trace != null) result.trace = trace;
    if (items != null) result.items.addAll(items);
    return result;
  }

  GetTraceDetailRes._();

  factory GetTraceDetailRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetTraceDetailRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTraceDetailRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<Trace>(1, _omitFieldNames ? '' : 'trace', subBuilder: Trace.create)
    ..pPM<TraceItem>(2, _omitFieldNames ? '' : 'items',
        subBuilder: TraceItem.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTraceDetailRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTraceDetailRes copyWith(void Function(GetTraceDetailRes) updates) =>
      super.copyWith((message) => updates(message as GetTraceDetailRes))
          as GetTraceDetailRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTraceDetailRes create() => GetTraceDetailRes._();
  @$core.override
  GetTraceDetailRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetTraceDetailRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTraceDetailRes>(create);
  static GetTraceDetailRes? _defaultInstance;

  @$pb.TagNumber(1)
  Trace get trace => $_getN(0);
  @$pb.TagNumber(1)
  set trace(Trace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTrace() => $_has(0);
  @$pb.TagNumber(1)
  void clearTrace() => $_clearField(1);
  @$pb.TagNumber(1)
  Trace ensureTrace() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<TraceItem> get items => $_getList(1);
}

class GetRandomMomentsReq extends $pb.GeneratedMessage {
  factory GetRandomMomentsReq({
    $core.int? count,
  }) {
    final result = create();
    if (count != null) result.count = count;
    return result;
  }

  GetRandomMomentsReq._();

  factory GetRandomMomentsReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetRandomMomentsReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRandomMomentsReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'count')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRandomMomentsReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRandomMomentsReq copyWith(void Function(GetRandomMomentsReq) updates) =>
      super.copyWith((message) => updates(message as GetRandomMomentsReq))
          as GetRandomMomentsReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetRandomMomentsReq create() => GetRandomMomentsReq._();
  @$core.override
  GetRandomMomentsReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetRandomMomentsReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRandomMomentsReq>(create);
  static GetRandomMomentsReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get count => $_getIZ(0);
  @$pb.TagNumber(1)
  set count($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCount() => $_has(0);
  @$pb.TagNumber(1)
  void clearCount() => $_clearField(1);
}

class GetRandomMomentsRes extends $pb.GeneratedMessage {
  factory GetRandomMomentsRes({
    $core.Iterable<Moment>? moments,
  }) {
    final result = create();
    if (moments != null) result.moments.addAll(moments);
    return result;
  }

  GetRandomMomentsRes._();

  factory GetRandomMomentsRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetRandomMomentsRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRandomMomentsRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..pPM<Moment>(1, _omitFieldNames ? '' : 'moments',
        subBuilder: Moment.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRandomMomentsRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRandomMomentsRes copyWith(void Function(GetRandomMomentsRes) updates) =>
      super.copyWith((message) => updates(message as GetRandomMomentsRes))
          as GetRandomMomentsRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetRandomMomentsRes create() => GetRandomMomentsRes._();
  @$core.override
  GetRandomMomentsRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetRandomMomentsRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRandomMomentsRes>(create);
  static GetRandomMomentsRes? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Moment> get moments => $_getList(0);
}

class StashTraceReq extends $pb.GeneratedMessage {
  factory StashTraceReq({
    $core.String? traceId,
  }) {
    final result = create();
    if (traceId != null) result.traceId = traceId;
    return result;
  }

  StashTraceReq._();

  factory StashTraceReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StashTraceReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StashTraceReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'traceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StashTraceReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StashTraceReq copyWith(void Function(StashTraceReq) updates) =>
      super.copyWith((message) => updates(message as StashTraceReq))
          as StashTraceReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StashTraceReq create() => StashTraceReq._();
  @$core.override
  StashTraceReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static StashTraceReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StashTraceReq>(create);
  static StashTraceReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get traceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set traceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTraceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTraceId() => $_clearField(1);
}

class StashTraceRes extends $pb.GeneratedMessage {
  factory StashTraceRes({
    Star? star,
  }) {
    final result = create();
    if (star != null) result.star = star;
    return result;
  }

  StashTraceRes._();

  factory StashTraceRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StashTraceRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StashTraceRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<Star>(1, _omitFieldNames ? '' : 'star', subBuilder: Star.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StashTraceRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StashTraceRes copyWith(void Function(StashTraceRes) updates) =>
      super.copyWith((message) => updates(message as StashTraceRes))
          as StashTraceRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StashTraceRes create() => StashTraceRes._();
  @$core.override
  StashTraceRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static StashTraceRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StashTraceRes>(create);
  static StashTraceRes? _defaultInstance;

  @$pb.TagNumber(1)
  Star get star => $_getN(0);
  @$pb.TagNumber(1)
  set star(Star value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStar() => $_has(0);
  @$pb.TagNumber(1)
  void clearStar() => $_clearField(1);
  @$pb.TagNumber(1)
  Star ensureStar() => $_ensure(0);
}

class ListConstellationsReq extends $pb.GeneratedMessage {
  factory ListConstellationsReq() => create();

  ListConstellationsReq._();

  factory ListConstellationsReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListConstellationsReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListConstellationsReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConstellationsReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConstellationsReq copyWith(
          void Function(ListConstellationsReq) updates) =>
      super.copyWith((message) => updates(message as ListConstellationsReq))
          as ListConstellationsReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListConstellationsReq create() => ListConstellationsReq._();
  @$core.override
  ListConstellationsReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListConstellationsReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListConstellationsReq>(create);
  static ListConstellationsReq? _defaultInstance;
}

class ListConstellationsRes extends $pb.GeneratedMessage {
  factory ListConstellationsRes({
    $core.Iterable<Constellation>? constellations,
    $core.int? totalStarCount,
  }) {
    final result = create();
    if (constellations != null) result.constellations.addAll(constellations);
    if (totalStarCount != null) result.totalStarCount = totalStarCount;
    return result;
  }

  ListConstellationsRes._();

  factory ListConstellationsRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListConstellationsRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListConstellationsRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..pPM<Constellation>(1, _omitFieldNames ? '' : 'constellations',
        subBuilder: Constellation.create)
    ..aI(2, _omitFieldNames ? '' : 'totalStarCount')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConstellationsRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConstellationsRes copyWith(
          void Function(ListConstellationsRes) updates) =>
      super.copyWith((message) => updates(message as ListConstellationsRes))
          as ListConstellationsRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListConstellationsRes create() => ListConstellationsRes._();
  @$core.override
  ListConstellationsRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListConstellationsRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListConstellationsRes>(create);
  static ListConstellationsRes? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Constellation> get constellations => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get totalStarCount => $_getIZ(1);
  @$pb.TagNumber(2)
  set totalStarCount($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotalStarCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotalStarCount() => $_clearField(2);
}

class GetConstellationReq extends $pb.GeneratedMessage {
  factory GetConstellationReq({
    $core.String? constellationId,
  }) {
    final result = create();
    if (constellationId != null) result.constellationId = constellationId;
    return result;
  }

  GetConstellationReq._();

  factory GetConstellationReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetConstellationReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetConstellationReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'constellationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetConstellationReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetConstellationReq copyWith(void Function(GetConstellationReq) updates) =>
      super.copyWith((message) => updates(message as GetConstellationReq))
          as GetConstellationReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetConstellationReq create() => GetConstellationReq._();
  @$core.override
  GetConstellationReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetConstellationReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetConstellationReq>(create);
  static GetConstellationReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get constellationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set constellationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasConstellationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearConstellationId() => $_clearField(1);
}

class GetConstellationRes extends $pb.GeneratedMessage {
  factory GetConstellationRes({
    Constellation? constellation,
    $core.Iterable<Moment>? moments,
    $core.Iterable<Star>? stars,
  }) {
    final result = create();
    if (constellation != null) result.constellation = constellation;
    if (moments != null) result.moments.addAll(moments);
    if (stars != null) result.stars.addAll(stars);
    return result;
  }

  GetConstellationRes._();

  factory GetConstellationRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetConstellationRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetConstellationRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<Constellation>(1, _omitFieldNames ? '' : 'constellation',
        subBuilder: Constellation.create)
    ..pPM<Moment>(2, _omitFieldNames ? '' : 'moments',
        subBuilder: Moment.create)
    ..pPM<Star>(3, _omitFieldNames ? '' : 'stars', subBuilder: Star.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetConstellationRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetConstellationRes copyWith(void Function(GetConstellationRes) updates) =>
      super.copyWith((message) => updates(message as GetConstellationRes))
          as GetConstellationRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetConstellationRes create() => GetConstellationRes._();
  @$core.override
  GetConstellationRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetConstellationRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetConstellationRes>(create);
  static GetConstellationRes? _defaultInstance;

  @$pb.TagNumber(1)
  Constellation get constellation => $_getN(0);
  @$pb.TagNumber(1)
  set constellation(Constellation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasConstellation() => $_has(0);
  @$pb.TagNumber(1)
  void clearConstellation() => $_clearField(1);
  @$pb.TagNumber(1)
  Constellation ensureConstellation() => $_ensure(0);

  /// ① ✦ 我发现（constellation_insight 已在 Constellation 消息中）
  /// ② 主题里说过的话（关联的所有 Moment）
  @$pb.TagNumber(2)
  $pb.PbList<Moment> get moments => $_getList(1);

  /// ③ 和那时的自己说说话（Star 列表，前端组装为 PastSelfCard）
  @$pb.TagNumber(3)
  $pb.PbList<Star> get stars => $_getList(2);
}

class StartChatReq extends $pb.GeneratedMessage {
  factory StartChatReq({
    $core.String? starId,
    $core.Iterable<$core.String>? contextMomentIds,
    $core.String? chatSessionId,
  }) {
    final result = create();
    if (starId != null) result.starId = starId;
    if (contextMomentIds != null)
      result.contextMomentIds.addAll(contextMomentIds);
    if (chatSessionId != null) result.chatSessionId = chatSessionId;
    return result;
  }

  StartChatReq._();

  factory StartChatReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StartChatReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartChatReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'starId')
    ..pPS(2, _omitFieldNames ? '' : 'contextMomentIds')
    ..aOS(3, _omitFieldNames ? '' : 'chatSessionId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartChatReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartChatReq copyWith(void Function(StartChatReq) updates) =>
      super.copyWith((message) => updates(message as StartChatReq))
          as StartChatReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartChatReq create() => StartChatReq._();
  @$core.override
  StartChatReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static StartChatReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartChatReq>(create);
  static StartChatReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get starId => $_getSZ(0);
  @$pb.TagNumber(1)
  set starId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasStarId() => $_has(0);
  @$pb.TagNumber(1)
  void clearStarId() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get contextMomentIds => $_getList(1);

  @$pb.TagNumber(3)
  $core.String get chatSessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set chatSessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasChatSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearChatSessionId() => $_clearField(3);
}

class StartChatRes extends $pb.GeneratedMessage {
  factory StartChatRes({
    $core.String? chatSessionId,
    ChatMessage? opening,
    $core.Iterable<ChatMessage>? history,
  }) {
    final result = create();
    if (chatSessionId != null) result.chatSessionId = chatSessionId;
    if (opening != null) result.opening = opening;
    if (history != null) result.history.addAll(history);
    return result;
  }

  StartChatRes._();

  factory StartChatRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StartChatRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartChatRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'chatSessionId')
    ..aOM<ChatMessage>(2, _omitFieldNames ? '' : 'opening',
        subBuilder: ChatMessage.create)
    ..pPM<ChatMessage>(3, _omitFieldNames ? '' : 'history',
        subBuilder: ChatMessage.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartChatRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartChatRes copyWith(void Function(StartChatRes) updates) =>
      super.copyWith((message) => updates(message as StartChatRes))
          as StartChatRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartChatRes create() => StartChatRes._();
  @$core.override
  StartChatRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static StartChatRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartChatRes>(create);
  static StartChatRes? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get chatSessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set chatSessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChatSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChatSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  ChatMessage get opening => $_getN(1);
  @$pb.TagNumber(2)
  set opening(ChatMessage value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOpening() => $_has(1);
  @$pb.TagNumber(2)
  void clearOpening() => $_clearField(2);
  @$pb.TagNumber(2)
  ChatMessage ensureOpening() => $_ensure(1);

  @$pb.TagNumber(3)
  $pb.PbList<ChatMessage> get history => $_getList(2);
}

class SendMessageReq extends $pb.GeneratedMessage {
  factory SendMessageReq({
    $core.String? chatSessionId,
    $core.String? content,
  }) {
    final result = create();
    if (chatSessionId != null) result.chatSessionId = chatSessionId;
    if (content != null) result.content = content;
    return result;
  }

  SendMessageReq._();

  factory SendMessageReq.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SendMessageReq.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendMessageReq',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'chatSessionId')
    ..aOS(2, _omitFieldNames ? '' : 'content')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendMessageReq clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendMessageReq copyWith(void Function(SendMessageReq) updates) =>
      super.copyWith((message) => updates(message as SendMessageReq))
          as SendMessageReq;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SendMessageReq create() => SendMessageReq._();
  @$core.override
  SendMessageReq createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static SendMessageReq getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SendMessageReq>(create);
  static SendMessageReq? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get chatSessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set chatSessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChatSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChatSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get content => $_getSZ(1);
  @$pb.TagNumber(2)
  set content($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearContent() => $_clearField(2);
}

class SendMessageRes extends $pb.GeneratedMessage {
  factory SendMessageRes({
    ChatMessage? reply,
  }) {
    final result = create();
    if (reply != null) result.reply = reply;
    return result;
  }

  SendMessageRes._();

  factory SendMessageRes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SendMessageRes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendMessageRes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ego'),
      createEmptyInstance: create)
    ..aOM<ChatMessage>(1, _omitFieldNames ? '' : 'reply',
        subBuilder: ChatMessage.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendMessageRes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendMessageRes copyWith(void Function(SendMessageRes) updates) =>
      super.copyWith((message) => updates(message as SendMessageRes))
          as SendMessageRes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SendMessageRes create() => SendMessageRes._();
  @$core.override
  SendMessageRes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static SendMessageRes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SendMessageRes>(create);
  static SendMessageRes? _defaultInstance;

  @$pb.TagNumber(1)
  ChatMessage get reply => $_getN(0);
  @$pb.TagNumber(1)
  set reply(ChatMessage value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasReply() => $_has(0);
  @$pb.TagNumber(1)
  void clearReply() => $_clearField(1);
  @$pb.TagNumber(1)
  ChatMessage ensureReply() => $_ensure(0);
}

class EgoApi {
  final $pb.RpcClient _client;

  EgoApi(this._client);

  /// ─── Auth（认证）─────────────────────────────────────
  $async.Future<LoginRes> login($pb.ClientContext? ctx, LoginReq request) =>
      _client.invoke<LoginRes>(ctx, 'Ego', 'Login', request, LoginRes());

  /// ─── Moment（话语）─────────────────────────────────────
  /// 写下一句话，保存 Moment + Echo，返回回声
  $async.Future<CreateMomentRes> createMoment(
          $pb.ClientContext? ctx, CreateMomentReq request) =>
      _client.invoke<CreateMomentRes>(
          ctx, 'Ego', 'CreateMoment', request, CreateMomentRes());

  /// 根据 ID 列表批量获取 Moment 内容
  $async.Future<GetMomentsRes> getMoments(
          $pb.ClientContext? ctx, GetMomentsReq request) =>
      _client.invoke<GetMomentsRes>(
          ctx, 'Ego', 'GetMoments', request, GetMomentsRes());

  /// 获取 AI 观察（per-Moment，结合 Echo）
  $async.Future<GenerateInsightRes> generateInsight(
          $pb.ClientContext? ctx, GenerateInsightReq request) =>
      _client.invoke<GenerateInsightRes>(
          ctx, 'Ego', 'GenerateInsight', request, GenerateInsightRes());

  /// ─── Trace（写作会话）─────────────────────────────────
  /// 获取过往 Trace 列表（游标分页，按月分组）
  $async.Future<ListTracesRes> listTraces(
          $pb.ClientContext? ctx, ListTracesReq request) =>
      _client.invoke<ListTracesRes>(
          ctx, 'Ego', 'ListTraces', request, ListTracesRes());

  /// 获取 Trace 详情（Item[] = <Moment, Echo[], Insight>）
  $async.Future<GetTraceDetailRes> getTraceDetail(
          $pb.ClientContext? ctx, GetTraceDetailReq request) =>
      _client.invoke<GetTraceDetailRes>(
          ctx, 'Ego', 'GetTraceDetail', request, GetTraceDetailRes());

  /// ─── Memory Dot（记忆光点盲盒）────────────────────────
  /// 随机获取 N 条历史话语
  $async.Future<GetRandomMomentsRes> getRandomMoments(
          $pb.ClientContext? ctx, GetRandomMomentsReq request) =>
      _client.invoke<GetRandomMomentsRes>(
          ctx, 'Ego', 'GetRandomMoments', request, GetRandomMomentsRes());

  /// ─── Stash（收进星图）─────────────────────────────────
  /// 将 Trace 寄存为 Star，触发异步聚类
  $async.Future<StashTraceRes> stashTrace(
          $pb.ClientContext? ctx, StashTraceReq request) =>
      _client.invoke<StashTraceRes>(
          ctx, 'Ego', 'StashTrace', request, StashTraceRes());

  /// ─── Constellation（星座）──────────────────────────────
  /// 获取所有星座列表（星图页渲染用）
  $async.Future<ListConstellationsRes> listConstellations(
          $pb.ClientContext? ctx, ListConstellationsReq request) =>
      _client.invoke<ListConstellationsRes>(
          ctx, 'Ego', 'ListConstellations', request, ListConstellationsRes());

  /// 获取星座详情（含 insight、Stars、topic_prompts）
  $async.Future<GetConstellationRes> getConstellation(
          $pb.ClientContext? ctx, GetConstellationReq request) =>
      _client.invoke<GetConstellationRes>(
          ctx, 'Ego', 'GetConstellation', request, GetConstellationRes());

  /// ─── Chat（和那时的自己说说话）──────────────────────────
  /// 开始一次对话（后端构建 past-self 上下文）
  $async.Future<StartChatRes> startChat(
          $pb.ClientContext? ctx, StartChatReq request) =>
      _client.invoke<StartChatRes>(
          ctx, 'Ego', 'StartChat', request, StartChatRes());

  /// 发送消息并获取回复
  $async.Future<SendMessageRes> sendMessage(
          $pb.ClientContext? ctx, SendMessageReq request) =>
      _client.invoke<SendMessageRes>(
          ctx, 'Ego', 'SendMessage', request, SendMessageRes());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');

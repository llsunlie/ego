// This is a generated file - do not edit.
//
// Generated from ego/api.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class ChatRole extends $pb.ProtobufEnum {
  static const ChatRole CHAT_ROLE_UNSPECIFIED =
      ChatRole._(0, _omitEnumNames ? '' : 'CHAT_ROLE_UNSPECIFIED');
  static const ChatRole USER = ChatRole._(1, _omitEnumNames ? '' : 'USER');
  static const ChatRole PAST_SELF =
      ChatRole._(2, _omitEnumNames ? '' : 'PAST_SELF');

  static const $core.List<ChatRole> values = <ChatRole>[
    CHAT_ROLE_UNSPECIFIED,
    USER,
    PAST_SELF,
  ];

  static final $core.List<ChatRole?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ChatRole? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ChatRole._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');

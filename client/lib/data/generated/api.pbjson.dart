// This is a generated file - do not edit.
//
// Generated from ego/api.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use chatRoleDescriptor instead')
const ChatRole$json = {
  '1': 'ChatRole',
  '2': [
    {'1': 'CHAT_ROLE_UNSPECIFIED', '2': 0},
    {'1': 'USER', '2': 1},
    {'1': 'PAST_SELF', '2': 2},
  ],
};

/// Descriptor for `ChatRole`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List chatRoleDescriptor = $convert.base64Decode(
    'CghDaGF0Um9sZRIZChVDSEFUX1JPTEVfVU5TUEVDSUZJRUQQABIICgRVU0VSEAESDQoJUEFTVF'
    '9TRUxGEAI=');

@$core.Deprecated('Use loginReqDescriptor instead')
const LoginReq$json = {
  '1': 'LoginReq',
  '2': [
    {'1': 'account', '3': 1, '4': 1, '5': 9, '10': 'account'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '10': 'password'},
  ],
};

/// Descriptor for `LoginReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginReqDescriptor = $convert.base64Decode(
    'CghMb2dpblJlcRIYCgdhY2NvdW50GAEgASgJUgdhY2NvdW50EhoKCHBhc3N3b3JkGAIgASgJUg'
    'hwYXNzd29yZA==');

@$core.Deprecated('Use loginResDescriptor instead')
const LoginRes$json = {
  '1': 'LoginRes',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {'1': 'created', '3': 2, '4': 1, '5': 8, '10': 'created'},
  ],
};

/// Descriptor for `LoginRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginResDescriptor = $convert.base64Decode(
    'CghMb2dpblJlcxIUCgV0b2tlbhgBIAEoCVIFdG9rZW4SGAoHY3JlYXRlZBgCIAEoCFIHY3JlYX'
    'RlZA==');

@$core.Deprecated('Use momentDescriptor instead')
const Moment$json = {
  '1': 'Moment',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'content', '3': 2, '4': 1, '5': 9, '10': 'content'},
    {'1': 'trace_id', '3': 3, '4': 1, '5': 9, '10': 'traceId'},
    {'1': 'created_at', '3': 4, '4': 1, '5': 3, '10': 'createdAt'},
  ],
};

/// Descriptor for `Moment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List momentDescriptor = $convert.base64Decode(
    'CgZNb21lbnQSDgoCaWQYASABKAlSAmlkEhgKB2NvbnRlbnQYAiABKAlSB2NvbnRlbnQSGQoIdH'
    'JhY2VfaWQYAyABKAlSB3RyYWNlSWQSHQoKY3JlYXRlZF9hdBgEIAEoA1IJY3JlYXRlZEF0');

@$core.Deprecated('Use echoDescriptor instead')
const Echo$json = {
  '1': 'Echo',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'moment_id', '3': 2, '4': 1, '5': 9, '10': 'momentId'},
    {
      '1': 'matched_moment_ids',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'matchedMomentIds'
    },
    {'1': 'similarities', '3': 4, '4': 3, '5': 2, '10': 'similarities'},
  ],
};

/// Descriptor for `Echo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List echoDescriptor = $convert.base64Decode(
    'CgRFY2hvEg4KAmlkGAEgASgJUgJpZBIbCgltb21lbnRfaWQYAiABKAlSCG1vbWVudElkEiwKEm'
    '1hdGNoZWRfbW9tZW50X2lkcxgDIAMoCVIQbWF0Y2hlZE1vbWVudElkcxIiCgxzaW1pbGFyaXRp'
    'ZXMYBCADKAJSDHNpbWlsYXJpdGllcw==');

@$core.Deprecated('Use insightDescriptor instead')
const Insight$json = {
  '1': 'Insight',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'moment_id', '3': 2, '4': 1, '5': 9, '10': 'momentId'},
    {'1': 'echo_id', '3': 3, '4': 1, '5': 9, '10': 'echoId'},
    {'1': 'text', '3': 4, '4': 1, '5': 9, '10': 'text'},
    {
      '1': 'related_moment_ids',
      '3': 5,
      '4': 3,
      '5': 9,
      '10': 'relatedMomentIds'
    },
  ],
};

/// Descriptor for `Insight`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List insightDescriptor = $convert.base64Decode(
    'CgdJbnNpZ2h0Eg4KAmlkGAEgASgJUgJpZBIbCgltb21lbnRfaWQYAiABKAlSCG1vbWVudElkEh'
    'cKB2VjaG9faWQYAyABKAlSBmVjaG9JZBISCgR0ZXh0GAQgASgJUgR0ZXh0EiwKEnJlbGF0ZWRf'
    'bW9tZW50X2lkcxgFIAMoCVIQcmVsYXRlZE1vbWVudElkcw==');

@$core.Deprecated('Use starDescriptor instead')
const Star$json = {
  '1': 'Star',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'trace_id', '3': 2, '4': 1, '5': 9, '10': 'traceId'},
    {'1': 'topic', '3': 3, '4': 1, '5': 9, '10': 'topic'},
  ],
};

/// Descriptor for `Star`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List starDescriptor = $convert.base64Decode(
    'CgRTdGFyEg4KAmlkGAEgASgJUgJpZBIZCgh0cmFjZV9pZBgCIAEoCVIHdHJhY2VJZBIUCgV0b3'
    'BpYxgDIAEoCVIFdG9waWM=');

@$core.Deprecated('Use constellationDescriptor instead')
const Constellation$json = {
  '1': 'Constellation',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'constellation_insight',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'constellationInsight'
    },
    {'1': 'star_ids', '3': 4, '4': 3, '5': 9, '10': 'starIds'},
    {'1': 'topic_prompts', '3': 5, '4': 3, '5': 9, '10': 'topicPrompts'},
    {'1': 'star_count', '3': 6, '4': 1, '5': 5, '10': 'starCount'},
    {'1': 'created_at', '3': 7, '4': 1, '5': 3, '10': 'createdAt'},
    {'1': 'updated_at', '3': 8, '4': 1, '5': 3, '10': 'updatedAt'},
  ],
};

/// Descriptor for `Constellation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List constellationDescriptor = $convert.base64Decode(
    'Cg1Db25zdGVsbGF0aW9uEg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEjMKFW'
    'NvbnN0ZWxsYXRpb25faW5zaWdodBgDIAEoCVIUY29uc3RlbGxhdGlvbkluc2lnaHQSGQoIc3Rh'
    'cl9pZHMYBCADKAlSB3N0YXJJZHMSIwoNdG9waWNfcHJvbXB0cxgFIAMoCVIMdG9waWNQcm9tcH'
    'RzEh0KCnN0YXJfY291bnQYBiABKAVSCXN0YXJDb3VudBIdCgpjcmVhdGVkX2F0GAcgASgDUglj'
    'cmVhdGVkQXQSHQoKdXBkYXRlZF9hdBgIIAEoA1IJdXBkYXRlZEF0');

@$core.Deprecated('Use chatMessageDescriptor instead')
const ChatMessage$json = {
  '1': 'ChatMessage',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'role', '3': 2, '4': 1, '5': 14, '6': '.ego.ChatRole', '10': 'role'},
    {'1': 'content', '3': 3, '4': 1, '5': 9, '10': 'content'},
    {
      '1': 'referenced',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.ego.MomentReference',
      '10': 'referenced'
    },
    {'1': 'timestamp', '3': 5, '4': 1, '5': 3, '10': 'timestamp'},
  ],
};

/// Descriptor for `ChatMessage`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatMessageDescriptor = $convert.base64Decode(
    'CgtDaGF0TWVzc2FnZRIOCgJpZBgBIAEoCVICaWQSIQoEcm9sZRgCIAEoDjINLmVnby5DaGF0Um'
    '9sZVIEcm9sZRIYCgdjb250ZW50GAMgASgJUgdjb250ZW50EjQKCnJlZmVyZW5jZWQYBCADKAsy'
    'FC5lZ28uTW9tZW50UmVmZXJlbmNlUgpyZWZlcmVuY2VkEhwKCXRpbWVzdGFtcBgFIAEoA1IJdG'
    'ltZXN0YW1w');

@$core.Deprecated('Use momentReferenceDescriptor instead')
const MomentReference$json = {
  '1': 'MomentReference',
  '2': [
    {'1': 'date', '3': 1, '4': 1, '5': 9, '10': 'date'},
    {'1': 'snippet', '3': 2, '4': 1, '5': 9, '10': 'snippet'},
  ],
};

/// Descriptor for `MomentReference`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List momentReferenceDescriptor = $convert.base64Decode(
    'Cg9Nb21lbnRSZWZlcmVuY2USEgoEZGF0ZRgBIAEoCVIEZGF0ZRIYCgdzbmlwcGV0GAIgASgJUg'
    'dzbmlwcGV0');

@$core.Deprecated('Use createMomentReqDescriptor instead')
const CreateMomentReq$json = {
  '1': 'CreateMomentReq',
  '2': [
    {'1': 'content', '3': 1, '4': 1, '5': 9, '10': 'content'},
    {'1': 'trace_id', '3': 2, '4': 1, '5': 9, '10': 'traceId'},
  ],
};

/// Descriptor for `CreateMomentReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMomentReqDescriptor = $convert.base64Decode(
    'Cg9DcmVhdGVNb21lbnRSZXESGAoHY29udGVudBgBIAEoCVIHY29udGVudBIZCgh0cmFjZV9pZB'
    'gCIAEoCVIHdHJhY2VJZA==');

@$core.Deprecated('Use createMomentResDescriptor instead')
const CreateMomentRes$json = {
  '1': 'CreateMomentRes',
  '2': [
    {
      '1': 'moment',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.ego.Moment',
      '10': 'moment'
    },
    {'1': 'echo', '3': 2, '4': 1, '5': 11, '6': '.ego.Echo', '10': 'echo'},
  ],
};

/// Descriptor for `CreateMomentRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMomentResDescriptor = $convert.base64Decode(
    'Cg9DcmVhdGVNb21lbnRSZXMSIwoGbW9tZW50GAEgASgLMgsuZWdvLk1vbWVudFIGbW9tZW50Eh'
    '0KBGVjaG8YAiABKAsyCS5lZ28uRWNob1IEZWNobw==');

@$core.Deprecated('Use getMomentsReqDescriptor instead')
const GetMomentsReq$json = {
  '1': 'GetMomentsReq',
  '2': [
    {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
  ],
};

/// Descriptor for `GetMomentsReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMomentsReqDescriptor =
    $convert.base64Decode('Cg1HZXRNb21lbnRzUmVxEhAKA2lkcxgBIAMoCVIDaWRz');

@$core.Deprecated('Use getMomentsResDescriptor instead')
const GetMomentsRes$json = {
  '1': 'GetMomentsRes',
  '2': [
    {
      '1': 'moments',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.ego.Moment',
      '10': 'moments'
    },
  ],
};

/// Descriptor for `GetMomentsRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMomentsResDescriptor = $convert.base64Decode(
    'Cg1HZXRNb21lbnRzUmVzEiUKB21vbWVudHMYASADKAsyCy5lZ28uTW9tZW50Ugdtb21lbnRz');

@$core.Deprecated('Use generateInsightReqDescriptor instead')
const GenerateInsightReq$json = {
  '1': 'GenerateInsightReq',
  '2': [
    {'1': 'moment_id', '3': 1, '4': 1, '5': 9, '10': 'momentId'},
    {'1': 'echo_id', '3': 2, '4': 1, '5': 9, '10': 'echoId'},
  ],
};

/// Descriptor for `GenerateInsightReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List generateInsightReqDescriptor = $convert.base64Decode(
    'ChJHZW5lcmF0ZUluc2lnaHRSZXESGwoJbW9tZW50X2lkGAEgASgJUghtb21lbnRJZBIXCgdlY2'
    'hvX2lkGAIgASgJUgZlY2hvSWQ=');

@$core.Deprecated('Use generateInsightResDescriptor instead')
const GenerateInsightRes$json = {
  '1': 'GenerateInsightRes',
  '2': [
    {
      '1': 'insight',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.ego.Insight',
      '10': 'insight'
    },
  ],
};

/// Descriptor for `GenerateInsightRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List generateInsightResDescriptor = $convert.base64Decode(
    'ChJHZW5lcmF0ZUluc2lnaHRSZXMSJgoHaW5zaWdodBgBIAEoCzIMLmVnby5JbnNpZ2h0Ugdpbn'
    'NpZ2h0');

@$core.Deprecated('Use traceDescriptor instead')
const Trace$json = {
  '1': 'Trace',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'motivation', '3': 2, '4': 1, '5': 9, '10': 'motivation'},
    {'1': 'stashed', '3': 3, '4': 1, '5': 8, '10': 'stashed'},
    {'1': 'created_at', '3': 4, '4': 1, '5': 3, '10': 'createdAt'},
  ],
};

/// Descriptor for `Trace`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List traceDescriptor = $convert.base64Decode(
    'CgVUcmFjZRIOCgJpZBgBIAEoCVICaWQSHgoKbW90aXZhdGlvbhgCIAEoCVIKbW90aXZhdGlvbh'
    'IYCgdzdGFzaGVkGAMgASgIUgdzdGFzaGVkEh0KCmNyZWF0ZWRfYXQYBCABKANSCWNyZWF0ZWRB'
    'dA==');

@$core.Deprecated('Use traceItemDescriptor instead')
const TraceItem$json = {
  '1': 'TraceItem',
  '2': [
    {
      '1': 'moment',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.ego.Moment',
      '10': 'moment'
    },
    {'1': 'echos', '3': 2, '4': 3, '5': 11, '6': '.ego.Echo', '10': 'echos'},
    {
      '1': 'insight',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.ego.Insight',
      '10': 'insight'
    },
  ],
};

/// Descriptor for `TraceItem`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List traceItemDescriptor = $convert.base64Decode(
    'CglUcmFjZUl0ZW0SIwoGbW9tZW50GAEgASgLMgsuZWdvLk1vbWVudFIGbW9tZW50Eh8KBWVjaG'
    '9zGAIgAygLMgkuZWdvLkVjaG9SBWVjaG9zEiYKB2luc2lnaHQYAyABKAsyDC5lZ28uSW5zaWdo'
    'dFIHaW5zaWdodA==');

@$core.Deprecated('Use listTracesReqDescriptor instead')
const ListTracesReq$json = {
  '1': 'ListTracesReq',
  '2': [
    {'1': 'cursor', '3': 1, '4': 1, '5': 9, '10': 'cursor'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListTracesReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTracesReqDescriptor = $convert.base64Decode(
    'Cg1MaXN0VHJhY2VzUmVxEhYKBmN1cnNvchgBIAEoCVIGY3Vyc29yEhsKCXBhZ2Vfc2l6ZRgCIA'
    'EoBVIIcGFnZVNpemU=');

@$core.Deprecated('Use listTracesResDescriptor instead')
const ListTracesRes$json = {
  '1': 'ListTracesRes',
  '2': [
    {'1': 'traces', '3': 1, '4': 3, '5': 11, '6': '.ego.Trace', '10': 'traces'},
    {'1': 'next_cursor', '3': 2, '4': 1, '5': 9, '10': 'nextCursor'},
    {'1': 'has_more', '3': 3, '4': 1, '5': 8, '10': 'hasMore'},
  ],
};

/// Descriptor for `ListTracesRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTracesResDescriptor = $convert.base64Decode(
    'Cg1MaXN0VHJhY2VzUmVzEiIKBnRyYWNlcxgBIAMoCzIKLmVnby5UcmFjZVIGdHJhY2VzEh8KC2'
    '5leHRfY3Vyc29yGAIgASgJUgpuZXh0Q3Vyc29yEhkKCGhhc19tb3JlGAMgASgIUgdoYXNNb3Jl');

@$core.Deprecated('Use getTraceDetailReqDescriptor instead')
const GetTraceDetailReq$json = {
  '1': 'GetTraceDetailReq',
  '2': [
    {'1': 'trace_id', '3': 1, '4': 1, '5': 9, '10': 'traceId'},
  ],
};

/// Descriptor for `GetTraceDetailReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTraceDetailReqDescriptor = $convert.base64Decode(
    'ChFHZXRUcmFjZURldGFpbFJlcRIZCgh0cmFjZV9pZBgBIAEoCVIHdHJhY2VJZA==');

@$core.Deprecated('Use getTraceDetailResDescriptor instead')
const GetTraceDetailRes$json = {
  '1': 'GetTraceDetailRes',
  '2': [
    {'1': 'trace', '3': 1, '4': 1, '5': 11, '6': '.ego.Trace', '10': 'trace'},
    {
      '1': 'items',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.ego.TraceItem',
      '10': 'items'
    },
  ],
};

/// Descriptor for `GetTraceDetailRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTraceDetailResDescriptor = $convert.base64Decode(
    'ChFHZXRUcmFjZURldGFpbFJlcxIgCgV0cmFjZRgBIAEoCzIKLmVnby5UcmFjZVIFdHJhY2USJA'
    'oFaXRlbXMYAiADKAsyDi5lZ28uVHJhY2VJdGVtUgVpdGVtcw==');

@$core.Deprecated('Use getRandomMomentsReqDescriptor instead')
const GetRandomMomentsReq$json = {
  '1': 'GetRandomMomentsReq',
  '2': [
    {'1': 'count', '3': 1, '4': 1, '5': 5, '10': 'count'},
  ],
};

/// Descriptor for `GetRandomMomentsReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRandomMomentsReqDescriptor =
    $convert.base64Decode(
        'ChNHZXRSYW5kb21Nb21lbnRzUmVxEhQKBWNvdW50GAEgASgFUgVjb3VudA==');

@$core.Deprecated('Use getRandomMomentsResDescriptor instead')
const GetRandomMomentsRes$json = {
  '1': 'GetRandomMomentsRes',
  '2': [
    {
      '1': 'moments',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.ego.Moment',
      '10': 'moments'
    },
  ],
};

/// Descriptor for `GetRandomMomentsRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRandomMomentsResDescriptor = $convert.base64Decode(
    'ChNHZXRSYW5kb21Nb21lbnRzUmVzEiUKB21vbWVudHMYASADKAsyCy5lZ28uTW9tZW50Ugdtb2'
    '1lbnRz');

@$core.Deprecated('Use stashTraceReqDescriptor instead')
const StashTraceReq$json = {
  '1': 'StashTraceReq',
  '2': [
    {'1': 'trace_id', '3': 1, '4': 1, '5': 9, '10': 'traceId'},
  ],
};

/// Descriptor for `StashTraceReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List stashTraceReqDescriptor = $convert
    .base64Decode('Cg1TdGFzaFRyYWNlUmVxEhkKCHRyYWNlX2lkGAEgASgJUgd0cmFjZUlk');

@$core.Deprecated('Use stashTraceResDescriptor instead')
const StashTraceRes$json = {
  '1': 'StashTraceRes',
  '2': [
    {'1': 'star', '3': 1, '4': 1, '5': 11, '6': '.ego.Star', '10': 'star'},
  ],
};

/// Descriptor for `StashTraceRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List stashTraceResDescriptor = $convert.base64Decode(
    'Cg1TdGFzaFRyYWNlUmVzEh0KBHN0YXIYASABKAsyCS5lZ28uU3RhclIEc3Rhcg==');

@$core.Deprecated('Use listConstellationsReqDescriptor instead')
const ListConstellationsReq$json = {
  '1': 'ListConstellationsReq',
};

/// Descriptor for `ListConstellationsReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listConstellationsReqDescriptor =
    $convert.base64Decode('ChVMaXN0Q29uc3RlbGxhdGlvbnNSZXE=');

@$core.Deprecated('Use listConstellationsResDescriptor instead')
const ListConstellationsRes$json = {
  '1': 'ListConstellationsRes',
  '2': [
    {
      '1': 'constellations',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.ego.Constellation',
      '10': 'constellations'
    },
    {'1': 'total_star_count', '3': 2, '4': 1, '5': 5, '10': 'totalStarCount'},
  ],
};

/// Descriptor for `ListConstellationsRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listConstellationsResDescriptor = $convert.base64Decode(
    'ChVMaXN0Q29uc3RlbGxhdGlvbnNSZXMSOgoOY29uc3RlbGxhdGlvbnMYASADKAsyEi5lZ28uQ2'
    '9uc3RlbGxhdGlvblIOY29uc3RlbGxhdGlvbnMSKAoQdG90YWxfc3Rhcl9jb3VudBgCIAEoBVIO'
    'dG90YWxTdGFyQ291bnQ=');

@$core.Deprecated('Use getConstellationReqDescriptor instead')
const GetConstellationReq$json = {
  '1': 'GetConstellationReq',
  '2': [
    {'1': 'constellation_id', '3': 1, '4': 1, '5': 9, '10': 'constellationId'},
  ],
};

/// Descriptor for `GetConstellationReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getConstellationReqDescriptor = $convert.base64Decode(
    'ChNHZXRDb25zdGVsbGF0aW9uUmVxEikKEGNvbnN0ZWxsYXRpb25faWQYASABKAlSD2NvbnN0ZW'
    'xsYXRpb25JZA==');

@$core.Deprecated('Use getConstellationResDescriptor instead')
const GetConstellationRes$json = {
  '1': 'GetConstellationRes',
  '2': [
    {
      '1': 'constellation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.ego.Constellation',
      '10': 'constellation'
    },
    {
      '1': 'moments',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.ego.Moment',
      '10': 'moments'
    },
    {'1': 'stars', '3': 3, '4': 3, '5': 11, '6': '.ego.Star', '10': 'stars'},
  ],
};

/// Descriptor for `GetConstellationRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getConstellationResDescriptor = $convert.base64Decode(
    'ChNHZXRDb25zdGVsbGF0aW9uUmVzEjgKDWNvbnN0ZWxsYXRpb24YASABKAsyEi5lZ28uQ29uc3'
    'RlbGxhdGlvblINY29uc3RlbGxhdGlvbhIlCgdtb21lbnRzGAIgAygLMgsuZWdvLk1vbWVudFIH'
    'bW9tZW50cxIfCgVzdGFycxgDIAMoCzIJLmVnby5TdGFyUgVzdGFycw==');

@$core.Deprecated('Use startChatReqDescriptor instead')
const StartChatReq$json = {
  '1': 'StartChatReq',
  '2': [
    {'1': 'star_id', '3': 1, '4': 1, '5': 9, '10': 'starId'},
    {
      '1': 'context_moment_ids',
      '3': 2,
      '4': 3,
      '5': 9,
      '10': 'contextMomentIds'
    },
    {'1': 'chat_session_id', '3': 3, '4': 1, '5': 9, '10': 'chatSessionId'},
  ],
};

/// Descriptor for `StartChatReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startChatReqDescriptor = $convert.base64Decode(
    'CgxTdGFydENoYXRSZXESFwoHc3Rhcl9pZBgBIAEoCVIGc3RhcklkEiwKEmNvbnRleHRfbW9tZW'
    '50X2lkcxgCIAMoCVIQY29udGV4dE1vbWVudElkcxImCg9jaGF0X3Nlc3Npb25faWQYAyABKAlS'
    'DWNoYXRTZXNzaW9uSWQ=');

@$core.Deprecated('Use startChatResDescriptor instead')
const StartChatRes$json = {
  '1': 'StartChatRes',
  '2': [
    {'1': 'chat_session_id', '3': 1, '4': 1, '5': 9, '10': 'chatSessionId'},
    {
      '1': 'opening',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.ego.ChatMessage',
      '10': 'opening'
    },
    {
      '1': 'history',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.ego.ChatMessage',
      '10': 'history'
    },
  ],
};

/// Descriptor for `StartChatRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startChatResDescriptor = $convert.base64Decode(
    'CgxTdGFydENoYXRSZXMSJgoPY2hhdF9zZXNzaW9uX2lkGAEgASgJUg1jaGF0U2Vzc2lvbklkEi'
    'oKB29wZW5pbmcYAiABKAsyEC5lZ28uQ2hhdE1lc3NhZ2VSB29wZW5pbmcSKgoHaGlzdG9yeRgD'
    'IAMoCzIQLmVnby5DaGF0TWVzc2FnZVIHaGlzdG9yeQ==');

@$core.Deprecated('Use sendMessageReqDescriptor instead')
const SendMessageReq$json = {
  '1': 'SendMessageReq',
  '2': [
    {'1': 'chat_session_id', '3': 1, '4': 1, '5': 9, '10': 'chatSessionId'},
    {'1': 'content', '3': 2, '4': 1, '5': 9, '10': 'content'},
  ],
};

/// Descriptor for `SendMessageReq`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendMessageReqDescriptor = $convert.base64Decode(
    'Cg5TZW5kTWVzc2FnZVJlcRImCg9jaGF0X3Nlc3Npb25faWQYASABKAlSDWNoYXRTZXNzaW9uSW'
    'QSGAoHY29udGVudBgCIAEoCVIHY29udGVudA==');

@$core.Deprecated('Use sendMessageResDescriptor instead')
const SendMessageRes$json = {
  '1': 'SendMessageRes',
  '2': [
    {
      '1': 'reply',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.ego.ChatMessage',
      '10': 'reply'
    },
  ],
};

/// Descriptor for `SendMessageRes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendMessageResDescriptor = $convert.base64Decode(
    'Cg5TZW5kTWVzc2FnZVJlcxImCgVyZXBseRgBIAEoCzIQLmVnby5DaGF0TWVzc2FnZVIFcmVwbH'
    'k=');

const $core.Map<$core.String, $core.dynamic> EgoServiceBase$json = {
  '1': 'Ego',
  '2': [
    {'1': 'Login', '2': '.ego.LoginReq', '3': '.ego.LoginRes'},
    {
      '1': 'CreateMoment',
      '2': '.ego.CreateMomentReq',
      '3': '.ego.CreateMomentRes'
    },
    {'1': 'GetMoments', '2': '.ego.GetMomentsReq', '3': '.ego.GetMomentsRes'},
    {
      '1': 'GenerateInsight',
      '2': '.ego.GenerateInsightReq',
      '3': '.ego.GenerateInsightRes'
    },
    {'1': 'ListTraces', '2': '.ego.ListTracesReq', '3': '.ego.ListTracesRes'},
    {
      '1': 'GetTraceDetail',
      '2': '.ego.GetTraceDetailReq',
      '3': '.ego.GetTraceDetailRes'
    },
    {
      '1': 'GetRandomMoments',
      '2': '.ego.GetRandomMomentsReq',
      '3': '.ego.GetRandomMomentsRes'
    },
    {'1': 'StashTrace', '2': '.ego.StashTraceReq', '3': '.ego.StashTraceRes'},
    {
      '1': 'ListConstellations',
      '2': '.ego.ListConstellationsReq',
      '3': '.ego.ListConstellationsRes'
    },
    {
      '1': 'GetConstellation',
      '2': '.ego.GetConstellationReq',
      '3': '.ego.GetConstellationRes'
    },
    {'1': 'StartChat', '2': '.ego.StartChatReq', '3': '.ego.StartChatRes'},
    {
      '1': 'SendMessage',
      '2': '.ego.SendMessageReq',
      '3': '.ego.SendMessageRes'
    },
  ],
};

@$core.Deprecated('Use egoServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    EgoServiceBase$messageJson = {
  '.ego.LoginReq': LoginReq$json,
  '.ego.LoginRes': LoginRes$json,
  '.ego.CreateMomentReq': CreateMomentReq$json,
  '.ego.CreateMomentRes': CreateMomentRes$json,
  '.ego.Moment': Moment$json,
  '.ego.Echo': Echo$json,
  '.ego.GetMomentsReq': GetMomentsReq$json,
  '.ego.GetMomentsRes': GetMomentsRes$json,
  '.ego.GenerateInsightReq': GenerateInsightReq$json,
  '.ego.GenerateInsightRes': GenerateInsightRes$json,
  '.ego.Insight': Insight$json,
  '.ego.ListTracesReq': ListTracesReq$json,
  '.ego.ListTracesRes': ListTracesRes$json,
  '.ego.Trace': Trace$json,
  '.ego.GetTraceDetailReq': GetTraceDetailReq$json,
  '.ego.GetTraceDetailRes': GetTraceDetailRes$json,
  '.ego.TraceItem': TraceItem$json,
  '.ego.GetRandomMomentsReq': GetRandomMomentsReq$json,
  '.ego.GetRandomMomentsRes': GetRandomMomentsRes$json,
  '.ego.StashTraceReq': StashTraceReq$json,
  '.ego.StashTraceRes': StashTraceRes$json,
  '.ego.Star': Star$json,
  '.ego.ListConstellationsReq': ListConstellationsReq$json,
  '.ego.ListConstellationsRes': ListConstellationsRes$json,
  '.ego.Constellation': Constellation$json,
  '.ego.GetConstellationReq': GetConstellationReq$json,
  '.ego.GetConstellationRes': GetConstellationRes$json,
  '.ego.StartChatReq': StartChatReq$json,
  '.ego.StartChatRes': StartChatRes$json,
  '.ego.ChatMessage': ChatMessage$json,
  '.ego.MomentReference': MomentReference$json,
  '.ego.SendMessageReq': SendMessageReq$json,
  '.ego.SendMessageRes': SendMessageRes$json,
};

/// Descriptor for `Ego`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List egoServiceDescriptor = $convert.base64Decode(
    'CgNFZ28SJQoFTG9naW4SDS5lZ28uTG9naW5SZXEaDS5lZ28uTG9naW5SZXMSOgoMQ3JlYXRlTW'
    '9tZW50EhQuZWdvLkNyZWF0ZU1vbWVudFJlcRoULmVnby5DcmVhdGVNb21lbnRSZXMSNAoKR2V0'
    'TW9tZW50cxISLmVnby5HZXRNb21lbnRzUmVxGhIuZWdvLkdldE1vbWVudHNSZXMSQwoPR2VuZX'
    'JhdGVJbnNpZ2h0EhcuZWdvLkdlbmVyYXRlSW5zaWdodFJlcRoXLmVnby5HZW5lcmF0ZUluc2ln'
    'aHRSZXMSNAoKTGlzdFRyYWNlcxISLmVnby5MaXN0VHJhY2VzUmVxGhIuZWdvLkxpc3RUcmFjZX'
    'NSZXMSQAoOR2V0VHJhY2VEZXRhaWwSFi5lZ28uR2V0VHJhY2VEZXRhaWxSZXEaFi5lZ28uR2V0'
    'VHJhY2VEZXRhaWxSZXMSRgoQR2V0UmFuZG9tTW9tZW50cxIYLmVnby5HZXRSYW5kb21Nb21lbn'
    'RzUmVxGhguZWdvLkdldFJhbmRvbU1vbWVudHNSZXMSNAoKU3Rhc2hUcmFjZRISLmVnby5TdGFz'
    'aFRyYWNlUmVxGhIuZWdvLlN0YXNoVHJhY2VSZXMSTAoSTGlzdENvbnN0ZWxsYXRpb25zEhouZW'
    'dvLkxpc3RDb25zdGVsbGF0aW9uc1JlcRoaLmVnby5MaXN0Q29uc3RlbGxhdGlvbnNSZXMSRgoQ'
    'R2V0Q29uc3RlbGxhdGlvbhIYLmVnby5HZXRDb25zdGVsbGF0aW9uUmVxGhguZWdvLkdldENvbn'
    'N0ZWxsYXRpb25SZXMSMQoJU3RhcnRDaGF0EhEuZWdvLlN0YXJ0Q2hhdFJlcRoRLmVnby5TdGFy'
    'dENoYXRSZXMSNwoLU2VuZE1lc3NhZ2USEy5lZ28uU2VuZE1lc3NhZ2VSZXEaEy5lZ28uU2VuZE'
    '1lc3NhZ2VSZXM=');

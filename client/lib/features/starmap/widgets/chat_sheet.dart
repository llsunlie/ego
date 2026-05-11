import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/colors.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../data/generated/api.pb.dart' as pb;
import '../../../data/services/ego_client.dart';
import '../../../data/services/interceptors/auth_interceptor.dart';

final _chatSessionProvider =
    StateProvider<Map<String, String>>((ref) => {});

class ChatSheet extends ConsumerStatefulWidget {
  final pb.Star star;

  const ChatSheet({super.key, required this.star});

  @override
  ConsumerState<ChatSheet> createState() => _ChatSheetState();
}

class _ChatSheetState extends ConsumerState<ChatSheet> {
  String? _sessionId;
  final List<pb.ChatMessage> _messages = [];
  bool _loading = true;
  String? _error;
  bool _sending = false;
  final _textCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();

  @override
  void initState() {
    super.initState();
    _startChat();
  }

  @override
  void dispose() {
    _textCtrl.dispose();
    _scrollCtrl.dispose();
    super.dispose();
  }

  Future<void> _startChat() async {
    try {
      final client = ref.read(EgoClient.provider);
      final token = ref.read(authProvider).token;
      final existing = ref.read(_chatSessionProvider)[widget.star.id];
      final res = await client.stub.startChat(
        pb.StartChatReq(
          starId: widget.star.id,
          chatSessionId: existing ?? '',
        ),
        options: authCallOptions(token),
      );
      setState(() {
        _sessionId = res.chatSessionId;
        ref.read(_chatSessionProvider.notifier).update(
              (m) => {...m, widget.star.id: res.chatSessionId},
            );
        if (res.hasOpening()) {
          final openingId = res.opening.id;
          final inHistory = res.history.any((h) => h.id == openingId);
          if (!inHistory) {
            _messages.add(res.opening);
          }
        }
        _messages.addAll(res.history);
        _loading = false;
      });
      _scrollToBottom();
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  Future<void> _sendMessage() async {
    final content = _textCtrl.text.trim();
    if (content.isEmpty || _sending || _sessionId == null) return;

    final userMsg = pb.ChatMessage(
      role: pb.ChatRole.USER,
      content: content,
    );
    setState(() {
      _messages.add(userMsg);
      _sending = true;
    });
    _textCtrl.clear();

    _scrollToBottom();

    try {
      final client = ref.read(EgoClient.provider);
      final token = ref.read(authProvider).token;
      final res = await client.stub.sendMessage(
        pb.SendMessageReq(
          chatSessionId: _sessionId!,
          content: content,
        ),
        options: authCallOptions(token),
      );
      setState(() {
        if (res.hasReply()) {
          _messages.add(res.reply);
        }
        _sending = false;
      });
      _scrollToBottom();
    } catch (_) {
      setState(() => _sending = false);
    }
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtrl.hasClients) {
        _scrollCtrl.animateTo(
          _scrollCtrl.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return DraggableScrollableSheet(
      initialChildSize: 0.85,
      minChildSize: 0.5,
      maxChildSize: 0.95,
      builder: (_, scrollCtrl) => Container(
        decoration: const BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
        ),
        child: Column(
          children: [
            _buildHeader(),
            const Divider(color: AppColors.surfaceLight, height: 1),
            Expanded(child: _buildBody()),
            _buildInputBar(),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 8),
      child: Column(
        children: [
          // Drag handle
          Container(
            margin: const EdgeInsets.only(bottom: 8),
            width: 36,
            height: 4,
            decoration: BoxDecoration(
              color: AppColors.textHint.withValues(alpha: 0.3),
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          // Title
          Row(
            children: [
              Container(
                width: 8,
                height: 8,
                decoration: const BoxDecoration(
                  shape: BoxShape.circle,
                  color: AppColors.softPurple,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  widget.star.topic,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: AppColors.textPrimary,
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBody() {
    if (_loading) {
      return const Center(
        child: CircularProgressIndicator(color: AppColors.gold),
      );
    }
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off, color: AppColors.textHint, size: 32),
            const SizedBox(height: 10),
            const Text('连接失败',
                style: TextStyle(color: AppColors.textHint, fontSize: 13)),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () {
                setState(() {
                  _loading = true;
                  _error = null;
                });
                _startChat();
              },
              child: const Text('重试',
                  style: TextStyle(color: AppColors.gold)),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollCtrl,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      itemCount: _messages.length,
      itemBuilder: (_, i) => _ChatBubble(message: _messages[i]),
    );
  }

  Widget _buildInputBar() {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 8, 8, 12),
      decoration: const BoxDecoration(
        border: Border(
          top: BorderSide(color: AppColors.surfaceLight, width: 0.5),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Expanded(
            child: Container(
              constraints: const BoxConstraints(maxHeight: 100),
              decoration: BoxDecoration(
                color: AppColors.surfaceLight.withValues(alpha: 0.4),
                borderRadius: BorderRadius.circular(20),
              ),
              child: TextField(
                controller: _textCtrl,
                maxLines: 3,
                minLines: 1,
                style: const TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 14,
                ),
                decoration: const InputDecoration(
                  hintText: '说点什么...',
                  hintStyle: TextStyle(
                    color: AppColors.textHint,
                    fontSize: 14,
                  ),
                  border: InputBorder.none,
                  contentPadding:
                      EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                ),
                textInputAction: TextInputAction.newline,
                onChanged: (_) => setState(() {}),
              ),
            ),
          ),
          const SizedBox(width: 8),
          GestureDetector(
            onTap: _sending || _textCtrl.text.trim().isEmpty ? null : _sendMessage,
            child: Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _sending || _textCtrl.text.trim().isEmpty
                    ? AppColors.gold.withValues(alpha: 0.2)
                    : AppColors.gold.withValues(alpha: 0.6),
              ),
              child: _sending
                  ? const Padding(
                      padding: EdgeInsets.all(10),
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: AppColors.warmGold,
                      ),
                    )
                  : const Icon(Icons.arrow_upward,
                      color: AppColors.warmGold, size: 20),
            ),
          ),
        ],
      ),
    );
  }
}

class _ChatBubble extends StatelessWidget {
  final pb.ChatMessage message;

  const _ChatBubble({required this.message});

  bool get _isPastSelf => message.role == pb.ChatRole.PAST_SELF;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        mainAxisAlignment:
            _isPastSelf ? MainAxisAlignment.start : MainAxisAlignment.end,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          if (!_isPastSelf) const SizedBox(width: 48),
          Flexible(
            child: Column(
              crossAxisAlignment: _isPastSelf
                  ? CrossAxisAlignment.start
                  : CrossAxisAlignment.end,
              children: [
                Container(
                  constraints: const BoxConstraints(maxWidth: 280),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 10,
                  ),
                  decoration: BoxDecoration(
                    color: _isPastSelf
                        ? AppColors.pastSelfBubble
                        : AppColors.userBubble,
                    borderRadius: BorderRadius.only(
                      topLeft: const Radius.circular(14),
                      topRight: const Radius.circular(14),
                      bottomLeft: _isPastSelf
                          ? const Radius.circular(4)
                          : const Radius.circular(14),
                      bottomRight: _isPastSelf
                          ? const Radius.circular(14)
                          : const Radius.circular(4),
                    ),
                  ),
                  child: Text(
                    message.content,
                    style: const TextStyle(
                      color: AppColors.textPrimary,
                      fontSize: 14,
                      height: 1.6,
                    ),
                  ),
                ),
                if (_isPastSelf && message.referenced.isNotEmpty)
                  _ReferencedMoments(refs: message.referenced.toList()),
              ],
            ),
          ),
          if (_isPastSelf) const SizedBox(width: 48),
        ],
      ),
    );
  }
}

class _ReferencedMoments extends StatefulWidget {
  final List<pb.MomentReference> refs;

  const _ReferencedMoments({required this.refs});

  @override
  State<_ReferencedMoments> createState() => _ReferencedMomentsState();
}

class _ReferencedMomentsState extends State<_ReferencedMoments> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final refs = widget.refs;
    if (refs.isEmpty) return const SizedBox.shrink();

    return Padding(
      padding: const EdgeInsets.only(top: 6),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          GestureDetector(
            onTap: () => setState(() => _expanded = !_expanded),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: AppColors.gold.withValues(alpha: 0.15),
                ),
                color: AppColors.gold.withValues(alpha: 0.05),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    '你曾经写下了这些',
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFFA89878),
                      height: 1.4,
                    ),
                  ),
                  const SizedBox(width: 4),
                  Icon(
                    _expanded ? Icons.expand_less : Icons.expand_more,
                    size: 14,
                    color: const Color(0xFFA89878),
                  ),
                ],
              ),
            ),
          ),
          if (_expanded)
            ...refs.map((ref) => _singleRef(ref)),
        ],
      ),
    );
  }

  Widget _singleRef(pb.MomentReference ref) {
    return Container(
      margin: const EdgeInsets.only(top: 4),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      width: double.infinity,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: AppColors.gold.withValues(alpha: 0.15),
        ),
        color: AppColors.gold.withValues(alpha: 0.05),
      ),
      child: Text(
        '${ref.date}: ${ref.snippet}',
        style: const TextStyle(
          fontSize: 11,
          color: Color(0xFFA89878),
          height: 1.4,
        ),
      ),
    );
  }
}

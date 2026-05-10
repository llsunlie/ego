import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/theme/colors.dart';
import '../../core/providers/auth_provider.dart';
import '../../data/generated/api.pb.dart' as pb;
import '../../data/services/ego_client.dart';
import '../../data/services/interceptors/auth_interceptor.dart';
import 'widgets/insight_section.dart';
import 'widgets/past_self_card.dart';
import 'widgets/topic_prompt_card.dart';

class ConstellationDetailPage extends ConsumerStatefulWidget {
  final String constellationId;

  const ConstellationDetailPage({super.key, required this.constellationId});

  @override
  ConsumerState<ConstellationDetailPage> createState() =>
      _ConstellationDetailPageState();
}

class _ConstellationDetailPageState
    extends ConsumerState<ConstellationDetailPage> {
  pb.Constellation? _constellation;
  List<pb.Moment> _moments = [];
  List<pb.Star> _stars = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final client = ref.read(EgoClient.provider);
      final token = ref.read(authProvider).token;
      final res = await client.stub.getConstellation(
        pb.GetConstellationReq(constellationId: widget.constellationId),
        options: authCallOptions(token),
      );
      setState(() {
        _constellation = res.constellation;
        _moments = res.moments;
        _stars = res.stars;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.darkBg,
      appBar: AppBar(
        backgroundColor: AppColors.darkBg,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: AppColors.textPrimary),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: Text(
          _constellation?.name ?? '星座详情',
          style: const TextStyle(
            color: AppColors.textPrimary,
            fontSize: 18,
            fontWeight: FontWeight.w500,
          ),
        ),
        centerTitle: true,
      ),
      body: _buildBody(),
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
            const Icon(Icons.cloud_off, color: AppColors.textHint, size: 40),
            const SizedBox(height: 12),
            Text(
              '加载失败',
              style: TextStyle(color: AppColors.textHint, fontSize: 14),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: _load,
              child: const Text('重试', style: TextStyle(color: AppColors.gold)),
            ),
          ],
        ),
      );
    }

    final c = _constellation!;
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 32),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ① ✦ 我发现
          InsightSection(
            insight: c.constellationInsight,
            moments: _moments,
          ),
          const SizedBox(height: 24),

          // ② 和那时的自己说说话
          _ChatSection(
            stars: _stars,
            constellationId: widget.constellationId,
          ),
          const SizedBox(height: 24),

          // ③ ✦ 我想和你聊聊
          if (c.topicPrompts.isNotEmpty)
            TopicPromptSection(prompts: c.topicPrompts),
        ],
      ),
    );
  }
}

class _ChatSection extends ConsumerStatefulWidget {
  final List<pb.Star> stars;
  final String constellationId;

  const _ChatSection({required this.stars, required this.constellationId});

  @override
  ConsumerState<_ChatSection> createState() => _ChatSectionState();
}

class _ChatSectionState extends ConsumerState<_ChatSection> {
  bool _expanded = false;

  static const _sectionTitle = '和那时的自己说说话';
  static const _maxVisible = 3;

  @override
  Widget build(BuildContext context) {
    final stars = widget.stars;
    if (stars.isEmpty) return const SizedBox.shrink();

    final visible = _expanded ? stars : stars.take(_maxVisible).toList();
    final hasMore = stars.length > _maxVisible;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.only(bottom: 12),
          child: Text(
            _sectionTitle,
            style: TextStyle(
              color: AppColors.textSecondary,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
        ...visible.map(
          (star) => Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: PastSelfCard(star: star, constellationId: widget.constellationId),
          ),
        ),
        if (hasMore && !_expanded)
          GestureDetector(
            onTap: () => setState(() => _expanded = true),
            child: Container(
              padding: const EdgeInsets.symmetric(vertical: 10),
              alignment: Alignment.center,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    '…还有 ${stars.length - _maxVisible} 条',
                    style: const TextStyle(
                      color: AppColors.textHint,
                      fontSize: 12,
                    ),
                  ),
                  const SizedBox(width: 4),
                  const Icon(Icons.expand_more, color: AppColors.textHint, size: 16),
                ],
              ),
            ),
          ),
        if (hasMore && _expanded)
          GestureDetector(
            onTap: () => setState(() => _expanded = false),
            child: Container(
              padding: const EdgeInsets.symmetric(vertical: 10),
              alignment: Alignment.center,
              child: const Icon(Icons.expand_less, color: AppColors.textHint, size: 16),
            ),
          ),
      ],
    );
  }
}

import 'package:flutter/material.dart';
import '../../../core/theme/colors.dart';
import '../../../data/generated/api.pb.dart' as pb;

class InsightSection extends StatefulWidget {
  final String insight;
  final List<pb.Moment> moments;

  const InsightSection({
    super.key,
    required this.insight,
    required this.moments,
  });

  @override
  State<InsightSection> createState() => _InsightSectionState();
}

class _InsightSectionState extends State<InsightSection> {
  bool _momentsExpanded = false;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // "✦ 我发现" header
        const Padding(
          padding: EdgeInsets.only(bottom: 12),
          child: Text(
            '✦ 我发现',
            style: TextStyle(
              color: AppColors.gold,
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),

        // Insight card
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                Color(0x1AD4A853),
                Color(0x0AD4A853),
              ],
            ),
            border: Border.all(
              color: AppColors.gold.withValues(alpha: 0.15),
            ),
          ),
          child: Text(
            widget.insight,
            style: const TextStyle(
              color: AppColors.textPrimary,
              fontSize: 14,
              height: 1.7,
            ),
          ),
        ),

        const SizedBox(height: 12),

        // Collapsible moments
        if (widget.moments.isNotEmpty)
          GestureDetector(
            onTap: () => setState(() => _momentsExpanded = !_momentsExpanded),
            child: Container(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Row(
                children: [
                  Icon(
                    _momentsExpanded
                        ? Icons.expand_less
                        : Icons.expand_more,
                    color: AppColors.textHint,
                    size: 18,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    '主题里说过的 ${widget.moments.length} 句话',
                    style: const TextStyle(
                      color: AppColors.textHint,
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
          ),

        AnimatedCrossFade(
          firstChild: const SizedBox.shrink(),
          secondChild: _MomentsList(moments: widget.moments),
          crossFadeState: _momentsExpanded
              ? CrossFadeState.showSecond
              : CrossFadeState.showFirst,
          duration: const Duration(milliseconds: 250),
        ),
      ],
    );
  }
}

class _MomentsList extends StatelessWidget {
  final List<pb.Moment> moments;

  const _MomentsList({required this.moments});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: moments.map((m) {
        final dt = DateTime.fromMillisecondsSinceEpoch(m.createdAt.toInt());
        final dateStr = '${dt.month}月${dt.day}日';
        return Container(
          width: double.infinity,
          margin: const EdgeInsets.only(bottom: 8),
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            color: Colors.white.withValues(alpha: 0.03),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                dateStr,
                style: const TextStyle(
                  color: AppColors.textHint,
                  fontSize: 11,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                m.content,
                style: const TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 13,
                  height: 1.6,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ],
          ),
        );
      }).toList(),
    );
  }
}

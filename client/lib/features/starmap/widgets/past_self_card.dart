import 'package:flutter/material.dart';
import '../../../core/theme/colors.dart';
import '../../../data/generated/api.pb.dart' as pb;

class PastSelfCard extends StatelessWidget {
  final pb.Star star;
  final String constellationId;

  const PastSelfCard({
    super.key,
    required this.star,
    required this.constellationId,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () => _openChat(context),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          color: AppColors.softPurple.withValues(alpha: 0.08),
          border: Border.all(
            color: AppColors.softPurple.withValues(alpha: 0.12),
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 6,
                  height: 6,
                  decoration: const BoxDecoration(
                    shape: BoxShape.circle,
                    color: AppColors.softPurple,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    star.topic,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: AppColors.textPrimary,
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            const Text(
              '和那时的自己说说话',
              style: TextStyle(
                color: AppColors.textHint,
                fontSize: 11,
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _openChat(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (_) => DraggableScrollableSheet(
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
              // Handle bar
              Container(
                margin: const EdgeInsets.only(top: 12, bottom: 8),
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: AppColors.textHint.withValues(alpha: 0.3),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              // Title
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                child: Row(
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
                        star.topic,
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
              ),
              const Divider(color: AppColors.surfaceLight, height: 1),
              // Placeholder chat content
              Expanded(
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.chat_bubble_outline,
                        size: 40,
                        color: AppColors.textHint.withValues(alpha: 0.4),
                      ),
                      const SizedBox(height: 12),
                      const Text(
                        '对话模式即将上线',
                        style: TextStyle(color: AppColors.textHint, fontSize: 14),
                      ),
                      const SizedBox(height: 4),
                      const Text(
                        '届时你可以和过去的自己交谈',
                        style: TextStyle(color: AppColors.textHint, fontSize: 12),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

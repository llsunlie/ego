import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/theme/colors.dart';
import '../../../core/providers/tab_provider.dart';
import '../providers/starmap_provider.dart';

class TopicPromptSection extends StatelessWidget {
  final List<String> prompts;

  const TopicPromptSection({super.key, required this.prompts});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.only(bottom: 12),
          child: Text(
            '✦ 我想和你聊聊',
            style: TextStyle(
              color: AppColors.gold,
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
        ...prompts.map(
          (prompt) => Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: _TopicPromptCard(prompt: prompt),
          ),
        ),
      ],
    );
  }
}

class _TopicPromptCard extends ConsumerWidget {
  final String prompt;

  const _TopicPromptCard({required this.prompt});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return GestureDetector(
      onTap: () {
        ref.read(pendingTopicPromptProvider.notifier).state = prompt;
        ref.read(tabProvider.notifier).setIndex(0);
        context.go('/now');
      },
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          gradient: const LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              Color(0x14D4A853),
              Color(0x08D4A853),
            ],
          ),
          border: Border.all(
            color: AppColors.gold.withValues(alpha: 0.12),
          ),
        ),
        child: Row(
          children: [
            const Icon(Icons.chevron_right, color: AppColors.gold, size: 18),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                prompt,
                style: const TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 13,
                  height: 1.5,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

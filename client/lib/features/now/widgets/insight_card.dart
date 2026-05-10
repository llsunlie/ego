import 'package:flutter/material.dart';
import '../../../core/theme/colors.dart';
import '../../../data/generated/api.pb.dart' as pb;

class InsightCard extends StatelessWidget {
  final pb.Insight insight;

  const InsightCard({super.key, required this.insight});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(22),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: AppColors.gold.withValues(alpha: 0.25),
        ),
        gradient: LinearGradient(
          colors: [
            AppColors.gold.withValues(alpha: 0.08),
            AppColors.softPurple.withValues(alpha: 0.05),
          ],
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '✦ 我发现',
            style: TextStyle(
              fontSize: 10,
              color: Color(0xFFCCA880),
              letterSpacing: 2.5,
              fontWeight: FontWeight.w200,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            insight.text,
            style: const TextStyle(
              fontSize: 14,
              color: Color(0xFFF0E6D2),
              height: 1.8,
              fontWeight: FontWeight.w300,
            ),
          ),
        ],
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:fixnum/fixnum.dart';
import '../../../data/generated/api.pb.dart' as pb;

class TraceItem extends StatefulWidget {
  final pb.Trace trace;
  final int index;

  const TraceItem({super.key, required this.trace, required this.index});

  @override
  State<TraceItem> createState() => _TraceItemState();
}

class _TraceItemState extends State<TraceItem> {
  bool _hovered = false;

  static const _dotColors = [
    [Color(0xFF78A0FF), Color(0xFF5064C8)],
    [Color(0xFFFFDC96), Color(0xFFC89650)],
    [Color(0xFFC8A0FF), Color(0xFF7850C8)],
    [Color(0xFFA0E6C8), Color(0xFF50B48C)],
    [Color(0xFFFFB4DC), Color(0xFFC878A0)],
  ];

  static String _formatDate(Int64 createdAt) {
    final dt = DateTime.fromMillisecondsSinceEpoch(createdAt.toInt());
    return '${dt.month}月${dt.day}日';
  }

  @override
  Widget build(BuildContext context) {
    final palette = _dotColors[widget.index % _dotColors.length];
    final dateStr = _formatDate(widget.trace.createdAt);
    final trace = widget.trace;

    return MouseRegion(
      onEnter: (_) => setState(() => _hovered = true),
      onExit: (_) => setState(() => _hovered = false),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () => context.push('/past/detail/${trace.id}'),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          margin: const EdgeInsets.only(bottom: 10),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: _hovered
                  ? Colors.white.withValues(alpha: 0.08)
                  : Colors.white.withValues(alpha: 0.04),
            ),
            color: _hovered
                ? Colors.white.withValues(alpha: 0.04)
                : Colors.white.withValues(alpha: 0.02),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: 12,
                height: 12,
                margin: const EdgeInsets.only(top: 4, right: 14),
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [palette[0], palette[1].withValues(alpha: 0.3)],
                  ),
                ),
              ),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      dateStr,
                      style: const TextStyle(
                        fontSize: 11,
                        color: Color(0xFF5A5A70),
                        letterSpacing: 1,
                      ),
                    ),
                    if (trace.firstMomentContent.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        trace.firstMomentContent,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFFC8C8D8),
                          height: 1.6,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              if (trace.stashed)
                const Padding(
                  padding: EdgeInsets.only(top: 2, left: 8),
                  child: Text(
                    '✦ 已联结',
                    style: TextStyle(
                      fontSize: 10,
                      color: Color(0xFFCCA880),
                    ),
                  ),
                ),
              const SizedBox(width: 4),
              const Padding(
                padding: EdgeInsets.only(top: 4),
                child: Icon(
                  Icons.chevron_right,
                  size: 16,
                  color: Color(0xFF5A5A70),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

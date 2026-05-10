import 'dart:math';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fixnum/fixnum.dart';
import '../../../data/generated/api.pb.dart' as pb;
import '../providers/now_page_provider.dart';

class MemoryDotGroup extends ConsumerWidget {
  final bool dimmed;

  const MemoryDotGroup({super.key, this.dimmed = false});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dotsAsync = ref.watch(memoryDotsProvider);

    return dotsAsync.when(
      data: (moments) {
        if (moments.isEmpty) return const SizedBox.shrink();
        return AnimatedOpacity(
          duration: const Duration(milliseconds: 500),
          opacity: dimmed ? 0.3 : 1.0,
          child: SizedBox.expand(
            child: Stack(
              children: List.generate(moments.length, (i) {
                return _MemoryDot(
                  moment: moments[i],
                  variant: i % 3,
                );
              }),
            ),
          ),
        );
      },
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }
}

class _MemoryDot extends StatefulWidget {
  final pb.Moment moment;
  final int variant;

  const _MemoryDot({required this.moment, required this.variant});

  @override
  State<_MemoryDot> createState() => _MemoryDotState();
}

class _MemoryDotState extends State<_MemoryDot>
    with TickerProviderStateMixin {
  late final AnimationController _floatCtrl;
  late final AnimationController _pulseCtrl;

  static const _colors = [
    [Color(0xFFFFE6B4), Color(0xFFFFC878)],
    [Color(0xFFB4D2FF), Color(0xFF8CB4F0)],
    [Color(0xFFE6C8FF), Color(0xFFBEA0E6)],
  ];

  // Positions relative to center area (like prototype: around the breathing light)
  static const _anchors = [
    Alignment(-0.5, -0.6),  // top-left area
    Alignment(0.55, 0.45),  // bottom-right area
    Alignment(0.5, -0.35),  // top-right area
  ];

  static const _floatDurations = [20, 25, 22];
  static const _pulseDurations = [3, 4, 3];

  @override
  void initState() {
    super.initState();
    _floatCtrl = AnimationController(
      vsync: this,
      duration: Duration(seconds: _floatDurations[widget.variant]),
    )..repeat();
    _pulseCtrl = AnimationController(
      vsync: this,
      duration: Duration(seconds: _pulseDurations[widget.variant]),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _floatCtrl.dispose();
    _pulseCtrl.dispose();
    super.dispose();
  }

  void _showEnvelope(BuildContext context) {
    showGeneralDialog(
      context: context,
      barrierColor: Colors.black.withValues(alpha: 0.6),
      barrierDismissible: true,
      barrierLabel: 'close',
      pageBuilder: (context, _, __) => _EnvelopeCard(moment: widget.moment),
      transitionBuilder: (context, anim, _, child) {
        final scale = TweenSequence<double>([
          TweenSequenceItem(tween: Tween(begin: 0.3, end: 1.05), weight: 0.6),
          TweenSequenceItem(tween: Tween(begin: 1.05, end: 1.0), weight: 0.4),
        ]).animate(CurvedAnimation(
          parent: anim,
          curve: Curves.easeOut,
        ));
        final rotation = TweenSequence<double>([
          TweenSequenceItem(tween: Tween(begin: -0.014, end: 0.003), weight: 0.5),
          TweenSequenceItem(tween: Tween(begin: 0.003, end: 0.0), weight: 0.5),
        ]).animate(CurvedAnimation(
          parent: anim,
          curve: Curves.easeOut,
        ));
        return ScaleTransition(
          scale: scale,
          child: RotationTransition(
            turns: rotation,
            child: FadeTransition(
              opacity: CurvedAnimation(
                parent: anim,
                curve: const Interval(0, 0.4, curve: Curves.easeIn),
              ),
              child: child,
            ),
          ),
        );
      },
      transitionDuration: const Duration(milliseconds: 800),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colors = _colors[widget.variant];
    final anchor = _anchors[widget.variant];

    return AnimatedBuilder(
      animation: Listenable.merge([_floatCtrl, _pulseCtrl]),
      builder: (_, __) {
        final floatX = sin(_floatCtrl.value * 2 * pi + widget.variant) * 16;
        final floatY = cos(_floatCtrl.value * 2 * pi + widget.variant * 1.7) * 12;
        final pulse = 0.6 + _pulseCtrl.value * 0.4;

        return Positioned.fill(
          child: Align(
            alignment: Alignment(
              anchor.x + floatX / 200,
              anchor.y + floatY / 300,
            ),
            child: GestureDetector(
              onTap: () => _showEnvelope(context),
              child: Container(
                width: 16,
                height: 16,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: colors[0].withValues(alpha: pulse),
                  boxShadow: [
                    BoxShadow(
                      color: colors[1].withValues(alpha: 0.5 * pulse),
                      blurRadius: 8,
                      spreadRadius: 2,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _EnvelopeCard extends StatelessWidget {
  final pb.Moment moment;

  const _EnvelopeCard({required this.moment});

  @override
  Widget build(BuildContext context) {
    final dateStr = _formatDate(moment.createdAt);

    return Center(
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 30),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Envelope flap — triangle pointing down, overlapping the body
              ClipPath(
                clipper: _EnvelopeFlapClipper(),
                child: Container(
                  width: 280,
                  height: 50,
                  decoration: const BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        Color(0xF23C321E),
                        Color(0xF22D2616),
                      ],
                    ),
                    border: Border(
                      bottom: BorderSide(
                        color: Color(0x26FFDC96),
                        width: 1,
                      ),
                    ),
                  ),
                  child: const Center(
                    child: Padding(
                      padding: EdgeInsets.only(top: 8),
                      child: Text(
                        '✦',
                        style: TextStyle(fontSize: 10, color: Color(0x80FFDC96)),
                      ),
                    ),
                  ),
                ),
              ),
              // Envelope body — overlaps the flap slightly
              Transform.translate(
                offset: const Offset(0, -2),
                child: Container(
                  width: 280,
                  padding: const EdgeInsets.all(22),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      colors: [
                        Color(0xF2282214),
                        Color(0xF21E1A0F),
                      ],
                    ),
                    borderRadius:
                        const BorderRadius.vertical(bottom: Radius.circular(14)),
                    border: Border.all(color: const Color(0x26FFDC96)),
                    boxShadow: const [
                      BoxShadow(
                        color: Color(0x4D000000),
                        blurRadius: 40,
                        offset: Offset(0, 10),
                      ),
                    ],
                  ),
                  child: Column(
                    children: [
                      Text(
                        dateStr,
                        style: const TextStyle(
                          fontSize: 10,
                          color: Color(0xFF8A7A60),
                          letterSpacing: 2,
                          fontWeight: FontWeight.w200,
                        ),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        moment.content,
                        textAlign: TextAlign.center,
                        style: const TextStyle(
                          fontSize: 15,
                          color: Color(0xFFF0E6D0),
                          height: 1.9,
                          fontWeight: FontWeight.w300,
                          fontStyle: FontStyle.italic,
                        ),
                      ),
                      const SizedBox(height: 18),
                      GestureDetector(
                        onTap: () => Navigator.of(context).pop(),
                        child: const Text(
                          '轻轻合上',
                          style: TextStyle(
                            fontSize: 11,
                            color: Color(0xFF8A7A60),
                            letterSpacing: 2,
                            fontWeight: FontWeight.w200,
                          ),
                        ),
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

  String _formatDate(Int64 createdAt) {
    final dt = DateTime.fromMillisecondsSinceEpoch(createdAt.toInt());
    return '${dt.month}月${dt.day}日 的你';
  }
}

class _EnvelopeFlapClipper extends CustomClipper<Path> {
  @override
  Path getClip(Size size) {
    return Path()
      ..moveTo(0, size.height)
      ..lineTo(size.width / 2, 0)
      ..lineTo(size.width, size.height)
      ..close();
  }

  @override
  bool shouldReclip(covariant CustomClipper<Path> oldClipper) => false;
}

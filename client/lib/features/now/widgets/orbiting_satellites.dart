import 'dart:math';
import 'package:flutter/material.dart';

class OrbitingSatellites extends StatefulWidget {
  final bool dimmed;

  const OrbitingSatellites({super.key, this.dimmed = false});

  @override
  State<OrbitingSatellites> createState() => _OrbitingSatellitesState();
}

class _OrbitingSatellitesState extends State<OrbitingSatellites>
    with TickerProviderStateMixin {
  late final AnimationController _ctrl1;
  late final AnimationController _ctrl2;

  @override
  void initState() {
    super.initState();
    _ctrl1 = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 30),
    )..repeat();
    _ctrl2 = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 45),
    )..repeat();
  }

  @override
  void dispose() {
    _ctrl1.dispose();
    _ctrl2.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedOpacity(
      duration: const Duration(milliseconds: 500),
      opacity: widget.dimmed ? 0.3 : 1.0,
      child: SizedBox(
        width: 340,
        height: 340,
        child: AnimatedBuilder(
          animation: Listenable.merge([_ctrl1, _ctrl2]),
          builder: (_, __) {
            return Stack(
              alignment: Alignment.center,
              children: [
                // Satellite 1 — warm gold, 30s orbit, radius 160px
                _Orbiter(
                  angle: _ctrl1.value * 2 * pi,
                  radius: 160,
                  color: const Color(0xFFFFDC96),
                  glowColor: const Color(0xFFFFB450),
                  size: 14,
                ),
                // Satellite 2 — cool green, 45s orbit, radius ~155px
                _Orbiter(
                  angle: _ctrl2.value * 2 * pi + pi * 0.7,
                  radius: 155,
                  color: const Color(0xFFC8FFDC),
                  glowColor: const Color(0xFF78C8A0),
                  size: 12,
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _Orbiter extends StatelessWidget {
  final double angle;
  final double radius;
  final Color color;
  final Color glowColor;
  final double size;

  const _Orbiter({
    required this.angle,
    required this.radius,
    required this.color,
    required this.glowColor,
    required this.size,
  });

  @override
  Widget build(BuildContext context) {
    return Transform.rotate(
      angle: angle,
      child: Transform.translate(
        offset: Offset(radius, 0),
        child: Transform.rotate(
          angle: -angle, // counter-rotate to stay upright
          child: Container(
            width: size,
            height: size,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: color,
              boxShadow: [
                BoxShadow(
                  color: glowColor.withValues(alpha: 0.6),
                  blurRadius: 10,
                  spreadRadius: 3,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

import 'dart:math';
import 'dart:ui' show Offset;
import '../../../data/generated/api.pb.dart' as pb;

class PlacedConstellation {
  final pb.Constellation constellation;
  final Offset center;
  final List<Offset> starPositions;
  final double twinklePhase;

  const PlacedConstellation({
    required this.constellation,
    required this.center,
    required this.starPositions,
    required this.twinklePhase,
  });
}

class StarPositionEngine {
  static const double worldWidth = 1500.0;
  static const double worldHeight = 1500.0;
  static const double margin = 120.0;
  static const double minConstellationDistance = 180.0;
  static const double hitRadius = 45.0;

  static const int _maxDisplayStars = 7;

  /// Spread radius depends on star count — wide enough for visible connection lines.
  static double _spreadFor(int count) {
    if (count <= 2) return 70.0;
    if (count <= 4) return 90.0;
    return 110.0;
  }

  /// Distribute constellation centers.
  /// First constellation is placed at world center; rest use hash-based
  /// placement with overlap avoidance.
  static List<PlacedConstellation> placeAll(
    List<pb.Constellation> constellations,
  ) {
    final placed = <PlacedConstellation>[];
    for (int i = 0; i < constellations.length; i++) {
      final c = constellations[i];
      final center = (i == 0)
          ? const Offset(worldWidth / 2, worldHeight / 2)
          : _placeConstellation(c, placed);

      // Limit to _maxDisplayStars, preserve the first N
      final displayIds = c.starIds.length > _maxDisplayStars
          ? c.starIds.sublist(0, _maxDisplayStars)
          : c.starIds;
      final count = displayIds.length;

      final starPositions = <Offset>[];
      for (int j = 0; j < count; j++) {
        starPositions.add(_starOffset(center, displayIds[j], j, count));
      }
      placed.add(PlacedConstellation(
        constellation: c,
        center: center,
        starPositions: starPositions,
        twinklePhase: _hashToDouble(c.id, 0, 2 * pi),
      ));
    }
    return placed;
  }

  /// Compute a non-overlapping position for a constellation.
  static Offset _placeConstellation(
    pb.Constellation c,
    List<PlacedConstellation> existing,
  ) {
    final seed = c.id.hashCode;
    var best = _seedToOffset(seed);
    double bestMinDist = 0;

    for (int attempt = 0; attempt < 20; attempt++) {
      final offset = (attempt == 0)
          ? best
          : _seedToOffset(seed ^ (attempt * 0x9E3779B9));
      double minDist = double.infinity;
      for (final pc in existing) {
        final d = (offset - pc.center).distance;
        if (d < minDist) minDist = d;
      }
      if (minDist >= minConstellationDistance) {
        return offset;
      }
      if (minDist > bestMinDist) {
        bestMinDist = minDist;
        best = offset;
      }
    }
    return best;
  }

  static Offset _seedToOffset(int seed) {
    final x = margin + _intHashToDouble(seed, 0, worldWidth - margin * 2);
    final y = margin + _intHashToDouble(~seed, 0, worldHeight - margin * 2);
    return Offset(x, y);
  }

  /// Structured star placement within a constellation — stars are distributed
  /// evenly around the center with a slight hash-based jitter for natural look.
  static Offset _starOffset(Offset center, String starId, int index, int total) {
    if (total == 1) return center;

    final spread = _spreadFor(total);
    final jitter = _hashToDouble(starId, -0.15, 0.15);

    if (total == 2) {
      // Two stars on opposite sides
      final dx = spread * (index == 0 ? -1 : 1);
      return center + Offset(dx + jitter * spread, jitter * spread * 0.6);
    }

    // 3+ stars: distribute evenly in a circle with jitter
    final angle = (2 * pi * index / total) + jitter;
    final radius = spread * (0.7 + 0.3 * _hashToDouble('$starId-r', 0, 1));
    return center + Offset(cos(angle) * radius, sin(angle) * radius);
  }

  /// Knuth multiplicative hash → double in [min, max).
  static double _hashToDouble(String s, double min, double max) {
    return _intHashToDouble(s.hashCode, min, max);
  }

  static double _intHashToDouble(int h, double min, double max) {
    final u = ((h & 0xFFFFFFFF) * 2654435761) & 0xFFFFFFFF;
    return min + (u / 0xFFFFFFFF) * (max - min);
  }
}

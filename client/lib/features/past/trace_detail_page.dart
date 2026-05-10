import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fixnum/fixnum.dart';
import '../../core/theme/colors.dart';
import '../../core/providers/auth_provider.dart';
import '../../data/generated/api.pb.dart' as pb;
import '../../data/services/ego_client.dart';
import '../../data/services/interceptors/auth_interceptor.dart';

class TraceDetailPage extends ConsumerStatefulWidget {
  final String traceId;

  const TraceDetailPage({super.key, required this.traceId});

  @override
  ConsumerState<TraceDetailPage> createState() => _TraceDetailPageState();
}

class _TraceDetailPageState extends ConsumerState<TraceDetailPage> {
  List<pb.TraceItem> _items = [];
  Map<String, pb.Moment> _matchedMoments = {};
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
      final res = await client.stub.getTraceDetail(
        pb.GetTraceDetailReq(traceId: widget.traceId),
        options: authCallOptions(token),
      );

      final items = res.items;

      // Collect all matched moment IDs across all echos
      final allIds = <String>{};
      for (final item in items) {
        for (final echo in item.echos) {
          allIds.addAll(echo.matchedMomentIds);
        }
      }

      // Fetch matched moments
      final moments = <String, pb.Moment>{};
      if (allIds.isNotEmpty) {
        final mRes = await client.stub.getMoments(
          pb.GetMomentsReq(ids: allIds.toList()),
          options: authCallOptions(token),
        );
        for (final m in mRes.moments) {
          moments[m.id] = m;
        }
      }

      setState(() {
        _items = items;
        _matchedMoments = moments;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  String _formatDate(Int64 createdAt) {
    final dt = DateTime.fromMillisecondsSinceEpoch(createdAt.toInt());
    return '${dt.month}月${dt.day}日 ${dt.hour}:${dt.minute.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0D0D14),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Color(0xFF8A8AA0)),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: const Text(
          '过往详情',
          style: TextStyle(
            color: Color(0xFF8A8AA0),
            fontSize: 14,
            fontWeight: FontWeight.w300,
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
        child: CircularProgressIndicator(
          color: Color(0xFFCCA880),
          strokeWidth: 2,
        ),
      );
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error!,
                style: const TextStyle(color: Color(0xFF5A5A70), fontSize: 14)),
            const SizedBox(height: 12),
            TextButton(
              onPressed: _load,
              child: const Text('重试',
                  style: TextStyle(color: Color(0xFFCCA880))),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      itemCount: _items.length,
      itemBuilder: (context, index) {
        final item = _items[index];
        final m = item.moment;
        return Padding(
          padding: const EdgeInsets.only(bottom: 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                _formatDate(m.createdAt),
                style: const TextStyle(
                  fontSize: 10,
                  color: Color(0xFF5A5A70),
                  letterSpacing: 1.5,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                m.content,
                style: const TextStyle(
                  fontSize: 15,
                  color: Color(0xFFD8D8E8),
                  height: 1.9,
                  fontWeight: FontWeight.w300,
                  fontStyle: FontStyle.italic,
                ),
              ),
              if (item.echos.isNotEmpty) ...[
                const SizedBox(height: 12),
                ...item.echos.map((echo) => _EchoCard(
                      echo: echo,
                      matchedMoments: _matchedMoments,
                    )),
              ],
              if (item.hasInsight() && item.insight.text.isNotEmpty) ...[
                const SizedBox(height: 4),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: AppColors.gold.withValues(alpha: 0.2),
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
                      const SizedBox(height: 10),
                      Text(
                        item.insight.text,
                        style: const TextStyle(
                          fontSize: 14,
                          color: Color(0xFFF0E6D2),
                          height: 1.8,
                          fontWeight: FontWeight.w300,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              if (index < _items.length - 1)
                Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: Container(
                    height: 1,
                    color: Colors.white.withValues(alpha: 0.04),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _EchoCard extends StatelessWidget {
  final pb.Echo echo;
  final Map<String, pb.Moment> matchedMoments;

  const _EchoCard({required this.echo, required this.matchedMoments});

  @override
  Widget build(BuildContext context) {
    final ids = echo.matchedMomentIds;
    final hasMatches = ids.isNotEmpty;
    final firstMatched = hasMatches ? matchedMoments[ids.first] : null;

    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withValues(alpha: 0.05)),
        color: Colors.white.withValues(alpha: 0.03),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '你之前也说过类似的',
            style: TextStyle(
              fontSize: 10,
              color: Color(0xFF5A5A70),
              letterSpacing: 1.5,
            ),
          ),
          const SizedBox(height: 10),
          if (firstMatched != null)
            Text(
              firstMatched.content,
              style: const TextStyle(
                fontSize: 14,
                color: Color(0xFFD0D0E0),
                height: 1.7,
                fontWeight: FontWeight.w300,
                fontStyle: FontStyle.italic,
              ),
            ),
          if (hasMatches && ids.length > 1) ...[
            const SizedBox(height: 8),
            _CandidateToggle(
              echo: echo,
              matchedMoments: matchedMoments,
            ),
          ],
        ],
      ),
    );
  }
}

class _CandidateToggle extends StatefulWidget {
  final pb.Echo echo;
  final Map<String, pb.Moment> matchedMoments;

  const _CandidateToggle({required this.echo, required this.matchedMoments});

  @override
  State<_CandidateToggle> createState() => _CandidateToggleState();
}

class _CandidateToggleState extends State<_CandidateToggle> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final ids = widget.echo.matchedMomentIds;
    final sims = widget.echo.similarities;

    return Column(
      children: [
        GestureDetector(
          onTap: () => setState(() => _expanded = !_expanded),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  _expanded ? '收起' : '之前的你还说过 ›',
                  style: const TextStyle(
                    fontSize: 11,
                    color: Color(0xFF7A7A90),
                    letterSpacing: 1,
                    fontWeight: FontWeight.w200,
                  ),
                ),
                const SizedBox(width: 4),
                AnimatedRotation(
                  turns: _expanded ? 0.5 : 0,
                  duration: const Duration(milliseconds: 200),
                  child: const Icon(
                    Icons.keyboard_arrow_down,
                    size: 14,
                    color: Color(0xFF7A7A90),
                  ),
                ),
              ],
            ),
          ),
        ),
        AnimatedSize(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
          alignment: Alignment.topCenter,
          child: _expanded
              ? Column(
                  children: List.generate(ids.length - 1, (i) {
                    final mi = i + 1; // skip first — already shown in _EchoCard
                    final sim = mi < sims.length
                        ? (sims[mi] * 100).toStringAsFixed(0)
                        : '--';
                    final moment = widget.matchedMoments[ids[mi]];
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Container(
                        width: double.infinity,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 12,
                        ),
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: Colors.white.withValues(alpha: 0.05),
                          ),
                          color: Colors.white.withValues(alpha: 0.02),
                        ),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Container(
                              width: 6,
                              height: 6,
                              margin: const EdgeInsets.only(top: 5),
                              decoration: BoxDecoration(
                                shape: BoxShape.circle,
                                color: AppColors.warmGold.withValues(
                                  alpha: 0.4 + i * 0.15,
                                ),
                              ),
                            ),
                            const SizedBox(width: 10),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  if (moment != null)
                                    Text(
                                      moment.content,
                                      style: const TextStyle(
                                        fontSize: 13,
                                        color: Color(0xFFB8B8C8),
                                        height: 1.6,
                                        fontWeight: FontWeight.w300,
                                        fontStyle: FontStyle.italic,
                                      ),
                                    ),
                                  const SizedBox(height: 6),
                                  Text(
                                    '$sim% 相似',
                                    style: const TextStyle(
                                      fontSize: 11,
                                      color: Color(0xFF6A6A80),
                                      fontWeight: FontWeight.w200,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  }),
                )
              : const SizedBox.shrink(),
        ),
      ],
    );
  }
}

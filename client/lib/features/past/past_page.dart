import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers/tab_provider.dart';
import 'providers/past_page_provider.dart';
import 'widgets/trace_item.dart';

class PastPage extends ConsumerStatefulWidget {
  const PastPage({super.key});

  @override
  ConsumerState<PastPage> createState() => _PastPageState();
}

class _PastPageState extends ConsumerState<PastPage> {
  final _scrollCtrl = ScrollController();
  bool _hasLoaded = false;

  @override
  void initState() {
    super.initState();
    _scrollCtrl.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollCtrl.removeListener(_onScroll);
    _scrollCtrl.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollCtrl.position.pixels >=
        _scrollCtrl.position.maxScrollExtent - 200) {
      ref.read(pastPageProvider.notifier).loadNextPage();
    }
  }

  @override
  Widget build(BuildContext context) {
    final tabIndex = ref.watch(tabProvider);
    final state = ref.watch(pastPageProvider);

    // Reload traces when switching to this tab
    if (tabIndex == 1 && !_hasLoaded && !state.isLoading) {
      _hasLoaded = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ref.read(pastPageProvider.notifier).loadFirstPage();
      });
    }
    if (tabIndex != 1) {
      _hasLoaded = false;
    }

    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            const _PageHeader(),
            Expanded(child: _buildBody(state)),
          ],
        ),
      ),
    );
  }

  Widget _buildBody(PastPageState state) {
    if (state.isLoading && state.traces.isEmpty) {
      return const Center(
        child: CircularProgressIndicator(
          color: Color(0xFFCCA880),
          strokeWidth: 2,
        ),
      );
    }

    if (state.error != null && state.traces.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('加载失败',
                style: TextStyle(color: Color(0xFF5A5A70), fontSize: 14)),
            const SizedBox(height: 12),
            TextButton(
              onPressed: () =>
                  ref.read(pastPageProvider.notifier).loadFirstPage(),
              child: const Text('重试',
                  style: TextStyle(color: Color(0xFFCCA880))),
            ),
          ],
        ),
      );
    }

    if (state.traces.isEmpty) {
      return const Center(
        child: Text('还没有写过什么',
            style: TextStyle(color: Color(0xFF5A5A70), fontSize: 14)),
      );
    }

    final sortedKeys = state.sortedMonthKeys;
    final groups = state.monthGroups;

    final flatItems = <Widget>[];
    for (final key in sortedKeys) {
      flatItems.add(_MonthLabel(key));
      for (var i = 0; i < groups[key]!.length; i++) {
        final trace = groups[key]![i];
        flatItems.add(TraceItem(trace: trace, index: i));
      }
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(pastPageProvider.notifier).refresh(),
      color: const Color(0xFFCCA880),
      child: ListView.builder(
        controller: _scrollCtrl,
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
        itemCount:
            state.isLoadingMore ? flatItems.length + 1 : flatItems.length,
        itemBuilder: (context, index) {
          if (index >= flatItems.length) {
            return const Padding(
              padding: EdgeInsets.symmetric(vertical: 20),
              child: Center(
                child: SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: Color(0xFFCCA880),
                  ),
                ),
              ),
            );
          }
          return flatItems[index];
        },
      ),
    );
  }
}

class _MonthLabel extends StatelessWidget {
  final String text;
  const _MonthLabel(this.text);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 20, bottom: 12),
      child: Text(
        text,
        textAlign: TextAlign.center,
        style: const TextStyle(
          fontSize: 11,
          color: Color(0xFF5A5A70),
          letterSpacing: 3,
        ),
      ),
    );
  }
}

class _PageHeader extends StatelessWidget {
  const _PageHeader();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 8),
      child: Text(
        '每一次说出口的，都留在这里',
        style: TextStyle(
          fontSize: 14,
          color: Colors.white.withValues(alpha: 0.5),
          fontWeight: FontWeight.w300,
          letterSpacing: 1.5,
        ),
      ),
    );
  }
}

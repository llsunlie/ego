import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/tab_provider.dart';

class AppShell extends ConsumerWidget {
  final StatefulNavigationShell navigationShell;

  const AppShell(this.navigationShell, {super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tabIndex = ref.watch(tabProvider);

    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: ClipRect(
        child: BackdropFilter(
          filter: ui.ImageFilter.blur(sigmaX: 12, sigmaY: 12),
          child: Container(
            decoration: BoxDecoration(
              color: const Color(0xFF0D0D14).withValues(alpha: 0.82),
              border: const Border(
                top: BorderSide(color: Color(0xFF1A1A24), width: 0.5),
              ),
            ),
            child: BottomNavigationBar(
              currentIndex: tabIndex,
              onTap: (index) {
                ref.read(tabProvider.notifier).setIndex(index);
                navigationShell.goBranch(index);
              },
              backgroundColor: Colors.transparent,
              elevation: 0,
              type: BottomNavigationBarType.fixed,
              selectedItemColor: const Color(0xFFCCA880),
              unselectedItemColor: const Color(0xFF5A5A70),
              selectedFontSize: 10,
              unselectedFontSize: 10,
              items: const [
                BottomNavigationBarItem(
                  icon: Icon(Icons.wb_sunny_outlined),
                  activeIcon: Icon(Icons.wb_sunny),
                  label: '此刻',
                ),
                BottomNavigationBarItem(
                  icon: Icon(Icons.history),
                  activeIcon: Icon(Icons.history_toggle_off),
                  label: '过往',
                ),
                BottomNavigationBarItem(
                  icon: Icon(Icons.auto_awesome),
                  activeIcon: Icon(Icons.auto_awesome),
                  label: '星图',
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

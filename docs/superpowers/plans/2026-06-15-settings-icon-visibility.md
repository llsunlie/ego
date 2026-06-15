# Settings Icon Visibility Fix

> **For agentic workers:** Single-task plan, no subagent needed.

**Goal:** 设置 icon 仅在 now_page、past_page、starmap_page 显示，不在 push 子页面（constellation_detail_page、trace_detail_page）显示。

**Architecture:** 在 `AppShell.build()` 中通过 `GoRouterState.of(context).uri.path` 判断当前路径，仅当路径为 `/now`、`/past`、`/starmap` 时渲染设置 icon。

**Tech Stack:** Flutter + go_router

---

### Task 1: 条件渲染设置 icon

**Files:**
- Modify: `client/lib/shared/widgets/app_shell.dart`

- [ ] **Step 1: 添加路由判断逻辑**

在 `build()` 方法中获取当前路由 path，使用 `GoRouterState.of(context).uri.path`，仅在三根路由下渲染设置 icon。

```dart
@override
Widget build(BuildContext context, WidgetRef ref) {
  final tabIndex = ref.watch(tabProvider);
  final currentPath = GoRouterState.of(context).uri.path;
  final showSettingsIcon = currentPath == '/now' ||
      currentPath == '/past' ||
      currentPath == '/starmap';

  return Scaffold(
    body: Stack(
      children: [
        navigationShell,
        if (showSettingsIcon)
          Positioned(
            top: MediaQuery.of(context).padding.top + 8,
            left: 4,
            child: IconButton(
              icon: const Icon(
                Icons.settings_outlined,
                color: Color(0xFF5A5A70),
                size: 22,
              ),
              onPressed: () => context.push('/setting'),
            ),
          ),
      ],
    ),
    // ... bottomNavigationBar 不变
  );
}
```

- [ ] **Step 2: Flutter 静态分析验证**

```bash
cd client && flutter analyze
```

Expected: 零 issue.

- [ ] **Step 3: 真机验证**

在真机上确认：
1. `/now` → 设置 icon 可见 ✓
2. `/past` → 设置 icon 可见 ✓
3. `/starmap` → 设置 icon 可见 ✓
4. `/past/detail/:traceId` → 设置 icon 不可见 ✓
5. `/starmap/detail/:constellationId` → 设置 icon 不可见 ✓

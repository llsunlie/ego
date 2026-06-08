# Setting 页面添加服务条款 & 隐私政策 + UI 美化 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Setting 页面「关于」区新增服务条款/隐私政策入口，并整体美化 UI（每行 icon + click-to-copy / push navigation）。

**Architecture:** 纯前端单文件重构 `setting_page.dart`。将现有 `_infoRow` 升级为带 icon、分割线、复制/导航交互的 `_settingRow`。保持现有 `ConsumerStatefulWidget` + `initState` 加载 profile 的数据流不变。

**Tech Stack:** Flutter + Riverpod + GoRouter

---

### Task 1: 重构 SettingPage UI（icon + 分区 + 交互）

**Files:**
- Modify: `client/lib/features/setting/setting_page.dart`

- [ ] **Step 1: 添加 imports 并存储原始手机号**

在文件顶部 import 区域添加 `flutter/services.dart`：

```dart
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';  // ← 新增，用于 Clipboard
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/theme/colors.dart';
import '../../core/version.dart';
import '../../data/services/ego_client.dart';
import '../../data/generated/api.pbgrpc.dart' as grpc;
```

在 `_SettingPageState` 中新增字段存储原始手机号：

```dart
class _SettingPageState extends ConsumerState<SettingPage> {
  grpc.GetProfileRes? _profile;
  bool _loading = true;
  String? _error;
  String _rawPhone = '';  // ← 新增：存储脱敏前的原始手机号
```

- [ ] **Step 2: 更新 `_loadProfile` 保存原始手机号**

找到 `_loadProfile` 方法，在 setState 同步保存 `_rawPhone`：

```dart
  Future<void> _loadProfile() async {
    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.getProfile(ref);
      setState(() {
        _profile = res;
        _rawPhone = res.phone;  // ← 新增
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }
```

- [ ] **Step 3: 替换整个 `build` 方法**

将当前 `build` 方法整体替换为以下实现（新增 section header 样式、icon 行、分割线、法律信息行）：

```dart
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.darkBg,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: AppColors.gold),
          onPressed: () => context.pop(),
        ),
        title: const Text(
          '设置',
          style: TextStyle(
            color: AppColors.gold,
            fontSize: 18,
            fontWeight: FontWeight.w500,
          ),
        ),
        centerTitle: true,
      ),
      body: _loading
          ? const Center(
              child: CircularProgressIndicator(color: AppColors.gold),
            )
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.error_outline,
                          color: AppColors.textHint, size: 48),
                      const SizedBox(height: 16),
                      const Text(
                        '加载失败',
                        style: TextStyle(color: AppColors.textHint),
                      ),
                    ],
                  ),
                )
              : Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 32),
                      _sectionHeader('账号信息'),
                      const SizedBox(height: 8),
                      _settingRow(
                        icon: Icons.phone_android_outlined,
                        label: '手机号',
                        value: _maskPhone(_rawPhone),
                        onTap: () => _copyToClipboard(_rawPhone, '手机号已复制'),
                      ),
                      _rowDivider(),
                      _settingRow(
                        icon: Icons.calendar_today_outlined,
                        label: '注册时间',
                        value: _formatDate(_profile!.createdAt.toInt()),
                        onTap: () => _copyToClipboard(
                            _formatDate(_profile!.createdAt.toInt()), '注册时间已复制'),
                      ),
                      const SizedBox(height: 32),
                      _sectionHeader('关于'),
                      const SizedBox(height: 8),
                      _settingRow(
                        icon: Icons.info_outline,
                        label: '版本',
                        value: appVersion,
                        onTap: () => _copyToClipboard(appVersion, '版本号已复制'),
                      ),
                      _rowDivider(),
                      _settingRow(
                        icon: Icons.description_outlined,
                        label: '服务条款',
                        showArrow: true,
                        onTap: () => context.push('/terms'),
                      ),
                      _rowDivider(),
                      _settingRow(
                        icon: Icons.shield_outlined,
                        label: '隐私政策',
                        showArrow: true,
                        onTap: () => context.push('/privacy'),
                      ),
                      const SizedBox(height: 48),
                      SizedBox(
                        width: double.infinity,
                        child: TextButton(
                          onPressed: _logout,
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                              side: const BorderSide(
                                color: Color(0xFFE53935),
                                width: 0.5,
                              ),
                            ),
                          ),
                          child: const Text(
                            '退出登录',
                            style: TextStyle(
                              color: Color(0xFFE53935),
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
    );
  }
```

- [ ] **Step 4: 替换/新增辅助 widget 方法**

删除旧的 `_infoRow` 方法，添加以下四个新方法到 `_SettingPageState` 中（放在 `_maskPhone` 方法之前或之后均可）：

```dart
  Widget _sectionHeader(String title) {
    return Text(
      title,
      style: const TextStyle(
        color: AppColors.textHint,
        fontSize: 13,
      ),
    );
  }

  Widget _settingRow({
    required IconData icon,
    required String label,
    String? value,
    bool showArrow = false,
    VoidCallback? onTap,
  }) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 12),
        child: Row(
          children: [
            Icon(icon, color: AppColors.gold, size: 20),
            const SizedBox(width: 12),
            Text(
              label,
              style: const TextStyle(
                color: AppColors.textSecondary,
                fontSize: 14,
              ),
            ),
            const Spacer(),
            if (value != null)
              Text(
                value,
                style: const TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 14,
                ),
              ),
            if (showArrow) ...[
              const SizedBox(width: 4),
              const Icon(Icons.chevron_right, color: AppColors.textHint, size: 20),
            ],
          ],
        ),
      ),
    );
  }

  Widget _rowDivider() {
    return const Padding(
      padding: EdgeInsets.only(left: 32), // icon(20) + gap(12) = 32
      child: Divider(
        color: AppColors.surfaceLight,
        height: 1,
        thickness: 1,
      ),
    );
  }

  void _copyToClipboard(String text, String message) {
    Clipboard.setData(ClipboardData(text: text));
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        duration: const Duration(seconds: 1),
        backgroundColor: AppColors.surface,
        behavior: SnackBarBehavior.floating,
      ),
    );
  }
```

- [ ] **Step 5: 删除旧的 `_infoRow` 方法**

确认已删除旧的 `_infoRow` 方法（被 Step 4 的新方法替代）。

- [ ] **Step 6: Flutter 静态分析**

```bash
cd client && flutter analyze
```
预期：零 issue。

- [ ] **Step 7: Commit**

```bash
git add client/lib/features/setting/setting_page.dart docs/superpowers/specs/2026-06-09-setting-terms-privacy-design.md docs/superpowers/plans/2026-06-09-setting-terms-privacy-plan.md
git commit -m "feat(setting): add terms/privacy entries, beautify UI with icons and copy-to-clipboard

- Add 服务条款 and 隐私政策 clickable rows in 关于 section
- Each row now has icon + click-to-copy or push navigation
- Section headers and row dividers for better visual grouping

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

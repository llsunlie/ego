import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/theme/colors.dart';
import '../../core/version.dart';
import '../../data/services/ego_client.dart';
import '../now/widgets/starry_background.dart';
import '../../data/generated/api.pbgrpc.dart' as grpc;

class SettingPage extends ConsumerStatefulWidget {
  const SettingPage({super.key});

  @override
  ConsumerState<SettingPage> createState() => _SettingPageState();
}

class _SettingPageState extends ConsumerState<SettingPage> {
  grpc.GetProfileRes? _profile;
  bool _loading = true;
  String? _error;
  String _rawPhone = '';

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.getProfile(ref);
      setState(() {
        _profile = res;
        _rawPhone = res.phone;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  void _logout() {
    ref.read(authProvider.notifier).logout();
    context.go('/login');
  }

  String _maskPhone(String phone) {
    if (phone.length < 7) return phone;
    return '${phone.substring(0, 3)}****${phone.substring(phone.length - 4)}';
  }

  String _formatDate(int unixMs) {
    final dt = DateTime.fromMillisecondsSinceEpoch(unixMs);
    return '${dt.year}/${dt.month.toString().padLeft(2, '0')}/${dt.day.toString().padLeft(2, '0')}';
  }

  void _copyToClipboard(String text, String message) {
    Clipboard.setData(ClipboardData(text: text));
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message, style: const TextStyle(color: Colors.white)),
        duration: const Duration(seconds: 1),
        backgroundColor: AppColors.surface,
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

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
      padding: EdgeInsets.only(left: 32),
      child: Divider(
        color: AppColors.surfaceLight,
        height: 1,
        thickness: 1,
      ),
    );
  }

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
      body: Stack(
        children: [
          const StarryBackground(),
          _loading
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
                      _rowDivider(),
                      _settingRow(
                        icon: Icons.feedback_outlined,
                        label: '用户反馈',
                        showArrow: true,
                        onTap: () => context.push('/feedback'),
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
        ],
      ),
    );
  }
}

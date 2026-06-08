import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/theme/colors.dart';
import '../../core/version.dart';
import '../../data/services/ego_client.dart';
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
                      const Text(
                        '账号信息',
                        style: TextStyle(
                          color: AppColors.textHint,
                          fontSize: 13,
                        ),
                      ),
                      const SizedBox(height: 16),
                      _infoRow(
                        '手机号',
                        _maskPhone(_profile!.phone),
                      ),
                      const SizedBox(height: 12),
                      _infoRow(
                        '注册时间',
                        _formatDate(_profile!.createdAt.toInt()),
                      ),
                      const SizedBox(height: 32),
                      const Text(
                        '关于',
                        style: TextStyle(
                          color: AppColors.textHint,
                          fontSize: 13,
                        ),
                      ),
                      const SizedBox(height: 16),
                      _infoRow('版本', appVersion),
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

  Widget _infoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: const TextStyle(
            color: AppColors.textHint,
            fontSize: 15,
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            color: AppColors.textPrimary,
            fontSize: 15,
          ),
        ),
      ],
    );
  }
}

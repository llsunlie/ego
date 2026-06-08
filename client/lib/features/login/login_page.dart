import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:grpc/grpc_or_grpcweb.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/theme/colors.dart';
import '../../data/services/ego_client.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _phoneCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  final _codeCtrl = TextEditingController();
  int _step = 0;
  bool _loading = false;
  String? _error;
  bool _agreedToTerms = false;
  int _countdown = 0;
  String? _codeSentPhone; // 缓存已发验证码的手机号，避免重复发送触发频率限制

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _passwordCtrl.dispose();
    _codeCtrl.dispose();
    super.dispose();
  }

  void _setError(String? msg) {
    if (mounted) setState(() => _error = msg);
  }

  Future<void> _checkPhone() async {
    final phone = _phoneCtrl.text.trim();
    if (phone.isEmpty) {
      _setError('请输入手机号');
      return;
    }
    if (!RegExp(r'^1[3-9]\d{9}$').hasMatch(phone)) {
      _setError('请输入正确的手机号');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.checkPhone(phone);

      if (!mounted) return;

      if (res.registered) {
        setState(() { _loading = false; _step = 1; });
      } else if (phone == _codeSentPhone) {
        // Same phone — already sent SMS, go directly to code+password form
        setState(() { _loading = false; _step = 2; _agreedToTerms = false; });
      } else {
        // New phone: auto-send SMS and go directly to code+password form
        try {
          await client.sendVerificationCode(phone);
          if (!mounted) return;
          setState(() { _loading = false; _step = 2; _countdown = 60; _codeSentPhone = phone; _agreedToTerms = false; });
          _startCountdown();
        } catch (_) {
          if (!mounted) return;
          setState(() => _loading = false);
          _setError('发送验证码失败，请稍后重试');
        }
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        _setError('网络错误，请稍后重试');
      }
    }
  }

  Future<void> _resendCode() async {
    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      await client.sendVerificationCode(_phoneCtrl.text.trim());
      if (!mounted) return;
      setState(() { _loading = false; _countdown = 60; _codeSentPhone = _phoneCtrl.text.trim(); });
      _startCountdown();
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
      _setError('发送验证码失败，请稍后重试');
    }
  }

  void _startCountdown() {
    Future.delayed(const Duration(seconds: 1), () {
      if (mounted && _countdown > 0) {
        setState(() => _countdown--);
        _startCountdown();
      }
    });
  }

  Future<void> _login() async {
    final password = _passwordCtrl.text;
    if (password.isEmpty) {
      _setError('请输入密码');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.login(_phoneCtrl.text.trim(), password);
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.unauthenticated) {
          _setError('密码错误');
        } else if (e.code == StatusCode.notFound) {
          _setError('用户不存在');
        } else {
          _setError('登录失败，请稍后重试');
        }
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
      _setError('登录失败，请稍后重试');
    }
  }

  Future<void> _register() async {
    final code = _codeCtrl.text.trim();
    final password = _passwordCtrl.text;

    if (code.isEmpty) {
      _setError('请输入验证码');
      return;
    }
    if (password.length < 6) {
      _setError('密码至少 6 位');
      return;
    }
    if (!_agreedToTerms) {
      _setError('请阅读并同意服务条款和隐私政策');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.register(
        phone: _phoneCtrl.text.trim(),
        code: code,
        password: password,
      );
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.unauthenticated) {
          _setError('验证码错误');
        } else if (e.code == StatusCode.alreadyExists) {
          _setError('该手机号已注册，请返回登录');
        } else {
          _setError('注册失败，请稍后重试');
        }
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
      _setError('注册失败，请稍后重试');
    }
  }

  void _backToStep0() {
    setState(() {
      _step = 0;
      _error = null;
      _passwordCtrl.clear();
      _codeCtrl.clear();
    });
  }

  Future<void> _goToStep3() async {
    final phone = _phoneCtrl.text.trim();
    if (phone == _codeSentPhone) {
      // SMS 已发至该手机号且倒计时仍在运行，直接进入重置密码页
      setState(() { _step = 3; });
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      await client.sendVerificationCode(phone);
      if (!mounted) return;
      setState(() { _loading = false; _step = 3; _countdown = 60; _codeSentPhone = phone; });
      _startCountdown();
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
      _setError('发送验证码失败，请稍后重试');
    }
  }

  Future<void> _resetPassword() async {
    final code = _codeCtrl.text.trim();
    final password = _passwordCtrl.text;

    if (code.isEmpty) {
      _setError('请输入验证码');
      return;
    }
    if (password.length < 6) {
      _setError('密码至少 6 位');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.resetPassword(
        phone: _phoneCtrl.text.trim(),
        code: code,
        newPassword: password,
      );
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        _setError(e.message ?? '重置密码失败，请稍后重试');
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
      _setError('重置密码失败，请稍后重试');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 120,
                  height: 120,
                  decoration: const BoxDecoration(shape: BoxShape.circle),
                  clipBehavior: Clip.antiAlias,
                  child: Image.asset('ego-logo.webp', fit: BoxFit.cover),
                ),
                const SizedBox(height: 48),

                if (_step != 0)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      _step == 1 ? '密码登录' : _step == 3 ? '重置密码' : '创建账号',
                      style: const TextStyle(
                        fontSize: 16,
                        color: Color(0xFFCCA880),
                        fontWeight: FontWeight.w300,
                        letterSpacing: 2,
                      ),
                    ),
                  ),

                if (_step != 0) ...[
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        _phoneCtrl.text.trim(),
                        style: const TextStyle(fontSize: 14, color: Color(0xFFA0A0B8)),
                      ),
                      const SizedBox(width: 8),
                      GestureDetector(
                        onTap: _loading ? null : _backToStep0,
                        child: const Text(
                          '修改',
                          style: TextStyle(fontSize: 12, color: Color(0xFFCCA880)),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),
                ],

                // Step 0: Phone input
                if (_step == 0)
                  TextField(
                    controller: _phoneCtrl,
                    keyboardType: TextInputType.phone,
                    inputFormatters: [
                      FilteringTextInputFormatter.digitsOnly,
                      LengthLimitingTextInputFormatter(11),
                    ],
                    decoration: const InputDecoration(
                      hintText: '请输入手机号',
                      prefixIcon: Icon(Icons.phone_android_outlined),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _checkPhone(),
                  ),

                // Step 1: Password login
                if (_step == 1) ...[
                  TextField(
                    controller: _passwordCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(
                      hintText: '请输入密码',
                      prefixIcon: Icon(Icons.lock_outline),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _login(),
                  ),
                  const SizedBox(height: 8),
                  Align(
                    alignment: Alignment.centerRight,
                    child: GestureDetector(
                      onTap: _loading ? null : _goToStep3,
                      child: const Text(
                        '忘记密码？',
                        style: TextStyle(fontSize: 12, color: Color(0xFFCCA880)),
                      ),
                    ),
                  ),
                ],

                // Step 2: Verification code + set password
                if (_step == 2) ...[
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _codeCtrl,
                          keyboardType: TextInputType.number,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                            LengthLimitingTextInputFormatter(6),
                          ],
                          decoration: const InputDecoration(
                            hintText: '验证码',
                            prefixIcon: Icon(Icons.sms_outlined),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      SizedBox(
                        width: 100,
                        child: TextButton(
                          onPressed: _countdown > 0 || _loading ? null : () => _resendCode(),
                          child: Text(
                            _countdown > 0 ? '${_countdown}s' : '重新发送',
                            style: const TextStyle(fontSize: 12),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _passwordCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(
                      hintText: '设置密码（至少 6 位）',
                      prefixIcon: Icon(Icons.lock_outline),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _register(),
                  ),
                  const SizedBox(height: 16),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        height: 24,
                        width: 24,
                        child: Checkbox(
                          value: _agreedToTerms,
                          onChanged: (v) => setState(() => _agreedToTerms = v ?? false),
                          activeColor: AppColors.gold,
                          side: const BorderSide(color: AppColors.textSecondary, width: 1.5),
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: RichText(
                          text: TextSpan(
                            style: const TextStyle(
                              fontSize: 13,
                              color: AppColors.textSecondary,
                              height: 1.5,
                            ),
                            children: [
                              const TextSpan(text: '我已阅读并同意'),
                              TextSpan(
                                text: '《服务条款》',
                                style: const TextStyle(color: AppColors.coldBlue),
                                recognizer: TapGestureRecognizer()
                                  ..onTap = () => context.push('/terms'),
                              ),
                              const TextSpan(text: ' 和 '),
                              TextSpan(
                                text: '《隐私政策》',
                                style: const TextStyle(color: AppColors.coldBlue),
                                recognizer: TapGestureRecognizer()
                                  ..onTap = () => context.push('/privacy'),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ],

                // Step 3: Forgot password — code + new password
                if (_step == 3) ...[
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _codeCtrl,
                          keyboardType: TextInputType.number,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                            LengthLimitingTextInputFormatter(6),
                          ],
                          decoration: const InputDecoration(
                            hintText: '验证码',
                            prefixIcon: Icon(Icons.sms_outlined),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      SizedBox(
                        width: 100,
                        child: TextButton(
                          onPressed: _countdown > 0 || _loading ? null : () => _resendCode(),
                          child: Text(
                            _countdown > 0 ? '${_countdown}s' : '重新发送',
                            style: const TextStyle(fontSize: 12),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _passwordCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(
                      hintText: '设置新密码（至少 6 位）',
                      prefixIcon: Icon(Icons.lock_outline),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _resetPassword(),
                  ),
                ],

                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(_error!, style: const TextStyle(color: Colors.redAccent)),
                ],

                const SizedBox(height: 32),

                if (_step == 0)
                  ElevatedButton(
                    onPressed: _loading ? null : _checkPhone,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('下一步', style: TextStyle(fontSize: 16)),
                  ),
                if (_step == 1)
                  ElevatedButton(
                    onPressed: _loading ? null : _login,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('登录', style: TextStyle(fontSize: 16)),
                  ),
                if (_step == 2)
                  ElevatedButton(
                    onPressed: (_loading || !_agreedToTerms) ? null : _register,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('注册', style: TextStyle(fontSize: 16)),
                  ),
                if (_step == 3)
                  ElevatedButton(
                    onPressed: _loading ? null : _resetPassword,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('重置密码', style: TextStyle(fontSize: 16)),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

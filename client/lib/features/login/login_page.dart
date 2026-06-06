import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers/auth_provider.dart';
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
  int _step = 0; // 0=手机号, 1=密码登录, 2=验证码注册
  bool _loading = false;
  String? _error;
  int _countdown = 0;

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

  Future<void> _sendCode() async {
    final phone = _phoneCtrl.text.trim();
    if (phone.isEmpty) {
      _setError('请输入手机号');
      return;
    }
    if (!RegExp(r'^1[3-9]\d{9}$').hasMatch(phone)) {
      _setError('请输入正确的手机号');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.sendVerificationCode(phone);

      if (mounted) {
        setState(() {
          _loading = false;
          _step = res.registered ? 1 : 2;
          _countdown = 60;
        });
        _startCountdown();
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        _setError('发送验证码失败，请稍后重试');
      }
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
    final phone = _phoneCtrl.text.trim();
    final password = _passwordCtrl.text;

    if (password.isEmpty) {
      _setError('请输入密码');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.login(phone, password);
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        final msg = e.toString();
        if (msg.contains('密码错误')) {
          _setError('密码错误');
        } else if (msg.contains('用户不存在')) {
          _setError('用户不存在');
        } else {
          _setError('登录失败，请稍后重试');
        }
      }
    }
  }

  Future<void> _register() async {
    final phone = _phoneCtrl.text.trim();
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

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.register(phone: phone, code: code, password: password);
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        final msg = e.toString();
        if (msg.contains('验证码')) {
          _setError('验证码错误');
        } else if (msg.contains('已注册')) {
          _setError('该手机号已注册，请返回登录');
        } else {
          _setError('注册失败，请稍后重试');
        }
      }
    }
  }

  void _backToStep0() {
    setState(() {
      _step = 0;
      _error = null;
      _countdown = 0;
      _passwordCtrl.clear();
      _codeCtrl.clear();
    });
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
                      _step == 1 ? '密码登录' : '创建账号',
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
                    onSubmitted: (_) => _sendCode(),
                  ),

                if (_step == 1)
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
                          onPressed: _countdown > 0 || _loading ? null : _sendCode,
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
                ],

                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(_error!, style: const TextStyle(color: Colors.redAccent)),
                ],

                const SizedBox(height: 32),

                if (_step == 0)
                  ElevatedButton(
                    onPressed: _loading ? null : _sendCode,
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
                    onPressed: _loading ? null : _register,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('注册', style: TextStyle(fontSize: 16)),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

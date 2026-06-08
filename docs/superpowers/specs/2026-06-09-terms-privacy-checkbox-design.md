# Register Agreement Checkbox — Design Spec

**Date:** 2026-06-09
**Status:** Design approved

## Overview

在注册界面（Step 2）增加「我已阅读并同意《服务条款》和《隐私政策》」checkbox，并将「服务条款」「隐私政策」做成可点击链接，打开应用内页面展示完整协议文本。

## Scope

纯前端变更，不影响 proto、后端或数据库。

## Files

| Action | File | Description |
|--------|------|-------------|
| Modify | `client/lib/features/login/login_page.dart` | Step 2 增加 checkbox + 协议文字，注册前校验 |
| Create | `client/lib/features/login/terms_page.dart` | 服务条款页面 |
| Create | `client/lib/features/login/privacy_page.dart` | 隐私政策页面 |
| Modify | `client/lib/core/router/router.dart` | 新增 `/terms` `/privacy` 路由，免登录访问 |

## Design

### login_page.dart 变更

**新增状态**：
```dart
bool _agreedToTerms = false;
```

**Step 2 UI 变更**（密码输入框之后、注册按钮之前插入）：
```dart
Row(
  children: [
    Checkbox(
      value: _agreedToTerms,
      onChanged: (v) => setState(() => _agreedToTerms = v ?? false),
    ),
    Expanded(
      child: RichText(
        text: TextSpan(
          style: defaultStyle,
          children: [
            TextSpan(text: '我已阅读并同意'),
            TextSpan(text: '《服务条款》', style: linkStyle, recognizer: TapGestureRecognizer()..onTap = () => context.push('/terms')),
            TextSpan(text: ' 和 '),
            TextSpan(text: '《隐私政策》', style: linkStyle, recognizer: TapGestureRecognizer()..onTap = () => context.push('/privacy')),
          ],
        ),
      ),
    ),
  ],
),
```

**注册按钮状态**：
```dart
onPressed: (_loading || !_agreedToTerms) ? null : _register,
```

**`_register()` 额外校验**：
```dart
if (!_agreedToTerms) {
  _setError('请阅读并同意服务条款和隐私政策');
  return;
}
```

### terms_page.dart & privacy_page.dart

- `Scaffold` + `AppBar`（标题 + 返回按钮）
- `SingleChildScrollView` + `Padding` 包裹协议正文
- 使用项目暗色主题风格
- 内容为标准中文应用服务条款/隐私政策模板文本

### router.dart 变更

新增两个顶级路由：
```dart
GoRoute(path: '/terms', builder: (context, state) => const TermsPage()),
GoRoute(path: '/privacy', builder: (context, state) => const PrivacyPage()),
```

**路由守卫**：`/terms` 和 `/privacy` 需加入免登录路由列表：
```dart
final isTermsRoute = state.matchedLocation == '/terms';
final isPrivacyRoute = state.matchedLocation == '/privacy';
// redirect 条件:
if (!loggedIn && !isLoginRoute && !isTermsRoute && !isPrivacyRoute) return '/login';
```

## Testing

- [ ] Checkbox 未勾选时注册按钮 disabled
- [ ] 勾选后可点击注册
- [ ] 点击「服务条款」跳转到 `/terms` 页面
- [ ] 点击「隐私政策」跳转到 `/privacy` 页面
- [ ] 协议页面可正常返回
- [ ] 未登录状态下可访问 `/terms` `/privacy`
- [ ] Step 0 → Step 2 切换时 checkbox 重置为未勾选

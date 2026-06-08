# Register Agreement Checkbox Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在注册 Step 2 增加「我已阅读并同意《服务条款》和《隐私政策》」checkbox，可点击跳转应用内协议页面。

**Architecture:** 纯前端变更 — 两个新页面 (terms_page, privacy_page) + 路由注册 + login_page Step 2 增加 checkbox 与校验。无后端/proto/DB 变更。

**Tech Stack:** Flutter, Riverpod, GoRouter

---

### Task 1: Create Terms of Service page

**Files:**
- Create: `client/lib/features/login/terms_page.dart`

- [ ] **Step 1: Write `terms_page.dart`**

```dart
import 'package:flutter/material.dart';
import '../../core/theme/colors.dart';

class TermsPage extends StatelessWidget {
  const TermsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('服务条款'),
        backgroundColor: AppColors.darkBg,
        foregroundColor: AppColors.textPrimary,
      ),
      backgroundColor: AppColors.darkBg,
      body: const SingleChildScrollView(
        padding: EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '服务条款',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 8),
            Text(
              '更新日期：2026年6月9日',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
            ),
            SizedBox(height: 16),
            _Section(
              title: '一、服务说明',
              content: 'EGO 是一款基于人工智能的个人成长与自我探索应用（以下简称"本服务"）。本服务通过对话、记录、反思等功能，帮助用户更好地认识自己、管理情绪、追踪成长轨迹。',
            ),
            _Section(
              title: '二、用户注册与账号',
              content: '2.1 您在使用本服务前需要注册账号，注册时需提供手机号码。\n\n'
                  '2.2 您应当提供真实、准确的注册信息，并妥善保管账号和密码。因您保管不善导致的账号被盗用或损失，由您自行承担。\n\n'
                  '2.3 您不得将账号转让、出借或授权他人使用。一个手机号仅可注册一个账号。',
            ),
            _Section(
              title: '三、用户行为规范',
              content: '3.1 您在使用本服务过程中，应遵守中华人民共和国相关法律法规。\n\n'
                  '3.2 您不得利用本服务从事以下活动：\n'
                  '（1）发布、传播违法或不良信息；\n'
                  '（2）干扰本服务的正常运行；\n'
                  '（3）利用技术手段破解、反向工程本服务；\n'
                  '（4）其他违反法律法规或侵犯他人合法权益的行为。\n\n'
                  '3.3 如发现用户存在违规行为，我们有权暂停或终止向您提供服务。',
            ),
            _Section(
              title: '四、知识产权',
              content: '4.1 本服务的所有内容，包括但不限于文字、图片、软件、界面设计等，其知识产权归 EGO 所有或已获得合法授权。\n\n'
                  '4.2 您在使用本服务过程中产生的个人数据归您所有。您授予我们在提供服务所必需的范围内使用这些数据的权利。\n\n'
                  '4.3 未经明确授权，您不得复制、修改、传播本服务的任何内容。',
            ),
            _Section(
              title: '五、免责声明',
              content: '5.1 本服务提供的 AI 对话内容仅供参考，不构成任何医疗、心理或法律建议。如有心理健康问题，请寻求专业帮助。\n\n'
                  '5.2 我们致力于提供稳定、安全的服务，但不对因不可抗力、系统维护、网络故障等原因导致的服务中断承担责任。\n\n'
                  '5.3 我们有权在必要时修改本服务条款，修改后的条款将在应用内公布。继续使用本服务即表示您同意修改后的条款。',
            ),
            _Section(
              title: '六、终止服务',
              content: '6.1 您可随时停止使用本服务。如需注销账号，请联系我们。\n\n'
                  '6.2 如您违反本服务条款，我们有权暂停或终止向您提供服务，并保留追究法律责任的权利。',
            ),
            _Section(
              title: '七、法律适用与争议解决',
              content: '7.1 本条款的订立、执行和解释及争议的解决均适用中华人民共和国法律。\n\n'
                  '7.2 因本条款引起的争议，双方应友好协商解决；协商不成的，任何一方均可向有管辖权的人民法院提起诉讼。',
            ),
            _Section(
              title: '八、联系我们',
              content: '如您对本服务条款有任何疑问，请通过应用内反馈功能联系我们。',
            ),
            SizedBox(height: 48),
          ],
        ),
      ),
    );
  }
}

class _Section extends StatelessWidget {
  final String title;
  final String content;

  const _Section({required this.title, required this.content});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            content,
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textSecondary,
              height: 1.6,
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add client/lib/features/login/terms_page.dart
git commit -m "feat(login): add Terms of Service page"
```

---

### Task 2: Create Privacy Policy page

**Files:**
- Create: `client/lib/features/login/privacy_page.dart`

- [ ] **Step 1: Write `privacy_page.dart`**

```dart
import 'package:flutter/material.dart';
import '../../core/theme/colors.dart';

class PrivacyPage extends StatelessWidget {
  const PrivacyPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('隐私政策'),
        backgroundColor: AppColors.darkBg,
        foregroundColor: AppColors.textPrimary,
      ),
      backgroundColor: AppColors.darkBg,
      body: const SingleChildScrollView(
        padding: EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '隐私政策',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 8),
            Text(
              '更新日期：2026年6月9日',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
            ),
            SizedBox(height: 16),
            _Section(
              title: '一、信息收集',
              content: '1.1 我们收集的信息包括：\n'
                  '（1）注册信息：手机号码、密码（加密存储）；\n'
                  '（2）使用数据：您在使用本服务过程中产生的对话记录、情绪记录、反思笔记等内容；\n'
                  '（3）设备信息：设备型号、操作系统版本等基础信息。\n\n'
                  '1.2 我们仅在提供本服务所必需的范围内收集上述信息。',
            ),
            _Section(
              title: '二、信息使用',
              content: '2.1 我们使用收集的信息用于以下目的：\n'
                  '（1）创建和管理您的账号；\n'
                  '（2）提供个性化的 AI 对话与自我探索体验；\n'
                  '（3）优化和改进本服务的功能与性能；\n'
                  '（4）保障账号安全和防范欺诈。\n\n'
                  '2.2 我们不会将您的个人信息用于上述目的之外的用途，除非获得您的明确同意或法律要求。',
            ),
            _Section(
              title: '三、信息存储与保护',
              content: '3.1 您的数据存储在位于中华人民共和国的服务器上。\n\n'
                  '3.2 我们采用业界通行的安全技术（包括数据加密传输、访问控制、安全审计等）保护您的个人信息。\n\n'
                  '3.3 您的密码采用 bcrypt 哈希算法加密存储，我们无法获知您的明文密码。',
            ),
            _Section(
              title: '四、信息共享与披露',
              content: '4.1 我们承诺不会向任何第三方出售、出租或交易您的个人信息。\n\n'
                  '4.2 在以下情况下，我们可能共享必要的信息：\n'
                  '（1）获得您的明确同意；\n'
                  '（2）为完成您所请求的服务（如短信验证码发送至电信运营商）；\n'
                  '（3）法律法规要求或政府机关依法要求。',
            ),
            _Section(
              title: '五、AI 数据处理',
              content: '5.1 本服务使用 AI 模型处理您的对话内容以生成回复。AI 处理过程遵循数据最小化原则，仅传输必要的上下文信息。\n\n'
                  '5.2 我们不会将您的个人对话数据用于 AI 模型的训练。\n\n'
                  '5.3 您的对话数据仅存储在您的个人账号下，其他用户无法访问。',
            ),
            _Section(
              title: '六、您的权利',
              content: '您对个人信息享有以下权利：\n'
                  '（1）访问权：您可以查看您的账号信息和使用数据；\n'
                  '（2）更正权：您可以更正不准确的个人信息；\n'
                  '（3）删除权：您可以删除部分或全部数据；\n'
                  '（4）注销权：您可以申请注销账号，我们将删除您的全部数据。\n\n'
                  '如需行使上述权利，请通过应用内反馈功能联系我们。',
            ),
            _Section(
              title: '七、未成年人保护',
              content: '7.1 本服务不面向 14 周岁以下的未成年人。\n\n'
                  '7.2 如您为 14 至 18 周岁的未成年人，请在监护人指导下使用本服务。\n\n'
                  '7.3 如我们发现误收集了未成年人的个人信息，将立即删除。',
            ),
            _Section(
              title: '八、政策更新',
              content: '8.1 我们可能适时更新本隐私政策，更新后的版本将在应用内公布。\n\n'
                  '8.2 重大变更我们将通过应用内通知或短信方式告知您。\n\n'
                  '8.3 继续使用本服务即表示您同意更新后的隐私政策。',
            ),
            _Section(
              title: '九、联系我们',
              content: '如您对本隐私政策有任何疑问或建议，请通过应用内反馈功能联系我们。',
            ),
            SizedBox(height: 48),
          ],
        ),
      ),
    );
  }
}

class _Section extends StatelessWidget {
  final String title;
  final String content;

  const _Section({required this.title, required this.content});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            content,
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textSecondary,
              height: 1.6,
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add client/lib/features/login/privacy_page.dart
git commit -m "feat(login): add Privacy Policy page"
```

---

### Task 3: Register routes

**Files:**
- Modify: `client/lib/core/router/router.dart`

- [ ] **Step 1: Add imports and routes, update redirect guard**

Add imports at top of `router.dart`:

```dart
import '../../features/login/terms_page.dart';
import '../../features/login/privacy_page.dart';
```

Add two new `GoRoute` entries **before** the `StatefulShellRoute.indexedStack` block, alongside `/login` and `/onboard` and `/setting`:

```dart
GoRoute(
  path: '/terms',
  builder: (context, state) => const TermsPage(),
),
GoRoute(
  path: '/privacy',
  builder: (context, state) => const PrivacyPage(),
),
```

Update the `redirect` function to allow unauthenticated access to `/terms` and `/privacy`:

```dart
redirect: (context, state) {
  final loggedIn = authState.isLoggedIn;
  final isLoginRoute = state.matchedLocation == '/login';
  final isOnboardingRoute = state.matchedLocation == '/onboard';
  final isTermsRoute = state.matchedLocation == '/terms';
  final isPrivacyRoute = state.matchedLocation == '/privacy';

  if (!loggedIn && !isLoginRoute && !isTermsRoute && !isPrivacyRoute) return '/login';
  if (loggedIn && isLoginRoute) {
    return onboardingDone ? '/now' : '/onboard';
  }
  if (loggedIn && !onboardingDone && !isOnboardingRoute) return '/onboard';
  if (loggedIn && onboardingDone && isOnboardingRoute) return '/now';
  return null;
},
```

The `go_router` package provides `context.push()` navigation, but since we're in the login page we need to use `GoRouter.of(context).push('/terms')` or use `context.push('/terms')` from `go_router`. The login page uses `context.push` via go_router's extension.

- [ ] **Step 2: Verify Flutter analysis passes**

```bash
cd client && flutter analyze
```

Expected: No issues found.

- [ ] **Step 3: Commit**

```bash
git add client/lib/core/router/router.dart
git commit -m "feat(router): add /terms and /privacy routes with unauthenticated access"
```

---

### Task 4: Add checkbox to login page

**Files:**
- Modify: `client/lib/features/login/login_page.dart`

- [ ] **Step 1: Add state variable**

Add after `String? _error;` (line 23):

```dart
bool _agreedToTerms = false;
```

- [ ] **Step 2: Reset checkbox when entering Step 2**

In `_checkPhone()`, when entering Step 2, add `_agreedToTerms = false;`:

In the block where `_step = 2` is set (for new phone: line 66, for same phone: line 60), add `_agreedToTerms = false;`:

For the "same phone" case (around line 59-60):
```dart
setState(() { _loading = false; _step = 2; _agreedToTerms = false; });
```

For the "new phone" case (around line 66):
```dart
setState(() { _loading = false; _step = 2; _countdown = 60; _codeSentPhone = phone; _agreedToTerms = false; });
```

- [ ] **Step 3: Add import for go_router**

Add at top of `login_page.dart` after other imports:

```dart
import 'package:go_router/go_router.dart';
```

- [ ] **Step 4: Add checkbox + agreement text UI**

Insert after the password TextField (line 387 `],`) and before the `],` that closes Step 2's `if` block, add:

```dart
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
```

Add the missing import for `AppColors` at top if not already present (check existing imports). Actually `login_page.dart` currently doesn't import `AppColors` — it uses hardcoded colors (`Color(0xFFCCA880)`, etc.). Let's also add:

```dart
import '../../core/theme/colors.dart';
```

Wait — reviewing the existing code, the page uses hardcoded `Color(...)` values. To maintain consistency with the existing style, let me use the same pattern with `AppColors` for the new code since that's the project standard per `theme.dart`.

- [ ] **Step 5: Update register button to disable when not agreed**

Change the Step 2 button's `onPressed` from:
```dart
onPressed: _loading ? null : _register,
```
to:
```dart
onPressed: (_loading || !_agreedToTerms) ? null : _register,
```

- [ ] **Step 6: Add validation in _register()**

Add after the password length check (after line 150):

```dart
if (!_agreedToTerms) {
  _setError('请阅读并同意服务条款和隐私政策');
  return;
}
```

- [ ] **Step 7: Verify Flutter analysis passes**

```bash
cd client && flutter analyze
```

Expected: No issues found.

- [ ] **Step 8: Commit**

```bash
git add client/lib/features/login/login_page.dart
git commit -m "feat(login): add terms/privacy agreement checkbox to registration step"
```

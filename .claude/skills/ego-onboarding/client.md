---
name: ego-cli-onboarding
description: 引导页 context — 路由 /onboard，新用户登录后的 5 步引导流程，含情绪选择、模拟日记、AI 洞察预览。
---

# ego-cli-onboarding

引导页 context。新用户首次登录后进入的 5 步引导流程。使用此 skill 后可直接讨论或修改引导流程。

## 路由

`/onboard` — 登录后若未完成引导则自动跳转。

## 核心文件

| 文件 | 说明 |
|------|------|
| `client/lib/features/onboarding/onboarding_page.dart` | 引导页主文件，含 5 个 Step 组件 (Intro/SelectFeeling/Diary/Insight/Preview) |
| `client/lib/features/onboarding/onboarding_data.dart` | 引导内容数据：4 种情绪场景 × 多条日记 + insight + preview 文案 |
| `client/lib/core/providers/onboarding_provider.dart` | 引导完成状态管理 |
| `client/lib/core/router/router.dart` | 路由守卫：未完成引导 → 强制 `/onboard` |
| `client/lib/core/theme/colors.dart` | 主题色定义 (AppColors) |

## 状态管理

- **onboardingCompleteProvider** (`StateNotifierProvider`)
  - `complete()` 标记引导完成，持久化到 LocalStore
  - Router 监听此 provider 决定是否放行到 `/now`

## 引导流程 (5 Steps)

```
Step 0 (Intro) → Step 1 (SelectFeeling) → Step 2 (Diary) → Step 3 (Insight) → Step 4 (Preview)
```

- **Step 0 - Intro**: Logo 呼吸动画 + 产品介绍文案 → 点击 "开始体验"
- **Step 1 - SelectFeeling**: 4 种情绪选项（"还不错但不踏实"、"有点累"、"正在选择"、"想记住好瞬间"）
- **Step 2 - Diary**: 展示选中场景的模拟日记，可切换多条 ("换一条看看") → 点击 "有点像我"
- **Step 3 - Insight**: 展示 ego 的洞察（"✦ 我发现"）+ 可选回复输入框 + "收进星图" 按钮（展示 tip 说明） → "开始体验"
- **Step 4 - Preview**: 7 行逐行动画预览文案 → "开始我的第一条" → 调用 `complete()`

## 数据模型

```dart
OnboardingGroup { diary: List<OnboardingDiary>, diary2: OnboardingDiary, insightFull: String }
OnboardingDiary { date: String, text: String }
```

## 动画

- `_breathCtrl` — Logo 呼吸动画 (3s 循环)
- `_previewCtrl` — Preview step 逐行出现动画 (16s)
- Step 间切换使用 `AnimatedSwitcher` (500ms ease)

## 共享 Widget

引导页内定义的所有 widget 都是私有的 (`_PrimaryButton`, `_GhostButton`, `_DiaryCard`, `_OutlineButton`, `_ConnectorLine`)，不是公共组件。

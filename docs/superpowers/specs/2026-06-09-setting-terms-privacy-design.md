# Design: Setting 页面添加服务条款 & 隐私政策 + UI 美化

**Date:** 2026-06-09
**Branch:** test
**Status:** approved

## 概述

Setting 页面当前仅有账号信息和版本，缺少法律信息入口。同时现有行样式简陋。本次将：
1. 在「关于」区域新增「服务条款」和「隐私政策」入口
2. 美化 setting 页面：分区 card 化，每行带 icon + 交互（click-to-copy 或 push）

## 页面结构（最终）

```
┌──────────────────────────────────┐
│          账号信息                 │  ← section header
│  📱 手机号        138****8888   │  ← click to copy raw phone
│  📅 注册时间      2025/01/15    │  ← click to copy date text
├──────────────────────────────────┤
│          关于                    │  ← section header
│  ℹ️ 版本          v1.2.3        │  ← click to copy version
│  📄 服务条款          >         │  ← push /terms
│  🛡️ 隐私政策          >         │  ← push /privacy
├──────────────────────────────────┤
│        [ 退出登录 ]              │
└──────────────────────────────────┘
```

## 交互规则

| 行 | Icon | 点击行为 |
|----|------|---------|
| 手机号 | Icons.phone_android_outlined | 复制脱敏前原始手机号到剪贴板，SnackBar 提示 "已复制" |
| 注册时间 | Icons.calendar_today_outlined | 复制日期文本到剪贴板，SnackBar 提示 "已复制" |
| 版本 | Icons.info_outline | 复制版本号到剪贴板，SnackBar 提示 "已复制" |
| 服务条款 | Icons.description_outlined | `context.push('/terms')` |
| 隐私政策 | Icons.shield_outlined | `context.push('/privacy')` |

## 样式

- **Icon**: `AppColors.gold`，size 20
- **Label**: `AppColors.textSecondary`（原 textHint），13-14px
- **Value**: `AppColors.textPrimary`，14px
- **右箭头**: 仅在有下页的行显示，`AppColors.textHint`，size 18
- **行间分割**: 1px `AppColors.surfaceLight`（左侧缩进 48px 对齐纯文本区域，不贯穿 icon）
- **Section header**: 灰色小字 `AppColors.textHint`，13px，复用现有样式
- **区域间分割**: 16px 间距（SizedBox），无实线分割线
- **行高**: min 48px，padding vertical 12px

## 数据保留

- `_profile` (GetProfileRes) 存储原始手机号，行上显示脱敏值，copy 时用原始值
- 版本号来自 `appVersion` 常量

## 涉及文件

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `client/lib/features/setting/setting_page.dart` | Modify | 重构 UI 结构，新增行 widget + copy 交互 + 法律信息行 |

## 不变更

- 无 proto 变更
- 无后端变更
- 无路由变更（`/terms` `/privacy` 已存在）
- 无数据库变更
- 无新文件创建
- 不迁移 terms_page.dart / privacy_page.dart（保持在 login feature 下）

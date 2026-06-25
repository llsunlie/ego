# 设置页增加公安备案号

**日期**: 2026-06-25
**类型**: 前端 UI 变更
**状态**: 已确认

## 背景

当前设置页底部仅展示 ICP 备案号（闽ICP备2026020313号），需增加公安备案号（粤公网安备44049002001272号）及备案图标。

## 设计

### 底部备案行布局

将当前单独的 ICP 备案号行替换为图标 + 公安备案号 + 分隔符 + ICP 备案号并排：

```
Copyright © 2026 Ego 工作室 保留所有权利
[图标] 粤公网安备44049002001272号  |  闽ICP备2026020313号
```

### 交互行为

| 元素 | 点击行为 |
|------|---------|
| 公安备案号（图标 + 文字） | 复制 `https://beian.mps.gov.cn/#/query/webSearch?code=44049002001272` 到剪贴板，提示「公安备案号链接已复制」 |
| ICP 备案号 | 保持不变，复制 `https://beian.miit.gov.cn/` 到剪贴板 |

### 变更文件

| 文件 | 变更 |
|------|------|
| `client/备案图标.png` → `client/beian-icon.png` | 重命名 |
| `client/pubspec.yaml` | 新增 `beian-icon.png` 到 assets |
| `client/lib/features/setting/setting_page.dart` | 重构底部备案行 |

### 不变更

- Proto: 无接口变更
- 后端: 无业务逻辑变更
- 数据库: 无 schema 变更

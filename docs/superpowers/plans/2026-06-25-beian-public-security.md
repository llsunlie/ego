# 设置页增加公安备案号 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在设置页底部将公安备案号（粤公网安备44049002001272号）与现有 ICP 备案号并排展示，点击分别复制对应链接。

**Architecture:** 纯前端 UI 变更，不涉及 Proto/后端/数据库。修改 setting_page.dart 底部备案行，将图标 + 公安备案号 + 分隔符 + ICP 备案号整合为一行 Row，各自有独立的 GestureDetector 复制对应链接。

**Tech Stack:** Flutter + Dart

## Global Constraints

- 遵循现有代码风格：AppColors.textHint 灰色小字、fontSize 12、SnackBar 浮窗提示
- 图标尺寸：16×18（等比缩放 36×40）
- 公安备案号链接：`https://beian.mps.gov.cn/#/query/webSearch?code=44049002001272`
- ICP 备案号链接：`https://beian.miit.gov.cn/`（保持不变）
- 一次 commit，不拆分

---

### Task 1: 重命名备案图标并注册 asset

**Files:**
- Rename: `client/备案图标.png` → `client/beian-icon.png`
- Modify: `client/pubspec.yaml`

- [ ] **Step 1: 重命名图标文件**

```bash
mv client/备案图标.png client/beian-icon.png
```

- [ ] **Step 2: 在 pubspec.yaml 中声明 asset**

```yaml
# 将当前 assets 段：
  assets:
    - ego-logo.webp

# 改为：
  assets:
    - ego-logo.webp
    - beian-icon.png
```

- [ ] **Step 3: Commit**

不单独 commit，等全部 task 完成后一起提交。

---

### Task 2: 重构设置页底部备案行

**Files:**
- Modify: `client/lib/features/setting/setting_page.dart`

- [ ] **Step 1: 新增公安备案号常量**

在 `icpRegistrationNumber` 常量下方（第 13 行后）新增：

```dart
const String publicSecurityRegistrationNumber = '粤公网安备44049002001272号';
```

- [ ] **Step 2: 替换底部备案行 UI**

将第 278-298 行的 ICP 备案号 Padding 块：

```dart
                      Padding(
                        padding: const EdgeInsets.only(bottom: 24),
                        child: Center(
                          child: GestureDetector(
                            onTap: () => _copyToClipboard(
                              'https://beian.miit.gov.cn/',
                              '备案号链接已复制',
                            ),
                            child: const Padding(
                              padding: EdgeInsets.all(8),
                              child: Text(
                                icpRegistrationNumber,
                                style: TextStyle(
                                  color: AppColors.textHint,
                                  fontSize: 12,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
```

替换为：

```dart
                      Padding(
                        padding: const EdgeInsets.only(bottom: 24),
                        child: Center(
                          child: Padding(
                            padding: const EdgeInsets.all(8),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                GestureDetector(
                                  onTap: () => _copyToClipboard(
                                    'https://beian.mps.gov.cn/#/query/webSearch?code=44049002001272',
                                    '公安备案号链接已复制',
                                  ),
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Image.asset(
                                        'beian-icon.png',
                                        width: 16,
                                        height: 18,
                                      ),
                                      const SizedBox(width: 4),
                                      const Text(
                                        publicSecurityRegistrationNumber,
                                        style: TextStyle(
                                          color: AppColors.textHint,
                                          fontSize: 12,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                const Text(
                                  '|',
                                  style: TextStyle(
                                    color: AppColors.textHint,
                                    fontSize: 12,
                                  ),
                                ),
                                const SizedBox(width: 8),
                                GestureDetector(
                                  onTap: () => _copyToClipboard(
                                    'https://beian.miit.gov.cn/',
                                    '备案号链接已复制',
                                  ),
                                  child: const Text(
                                    icpRegistrationNumber,
                                    style: TextStyle(
                                      color: AppColors.textHint,
                                      fontSize: 12,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
```

- [ ] **Step 3: Commit**

不单独 commit，等全部 task 完成后一起提交。

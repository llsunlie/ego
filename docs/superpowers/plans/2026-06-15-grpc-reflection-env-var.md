# gRPC Reflection 环境变量控制 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将无条件开启的 gRPC reflection 改为由 `GRPC_REFLECTION=true` 环境变量控制，默认关闭。

**Architecture:** 在 Config 结构体中新增 `GRPC_REFLECTION` 字符串字段，从环境变量加载；bootstrap/server.go 中根据该值条件注册 reflection。

**Tech Stack:** Go 1.x, gRPC reflection API

---

### Task 1: Config — 新增 GRPC_REFLECTION 字段

**Files:**
- Modify: `server/internal/config/config.go`

- [ ] **Step 1: 在 Config struct 中新增字段**

在 `Config` struct 的 `RateLimitBucketTTL` 字段之后（第 63 行之后），新增:

```go
// gRPC Reflection
GRPC_REFLECTION string
```

- [ ] **Step 2: 在 Load() 中加载环境变量**

在 `Load()` 函数的 `RateLimitBucketTTL` 行之后（第 116 行之后），新增:

```go
GRPC_REFLECTION: os.Getenv("GRPC_REFLECTION"),
```

- [ ] **Step 3: 验证编译**

```bash
cd server && go build ./internal/config/...
```

Expected: 编译通过，无错误。

---

### Task 2: Bootstrap — 条件注册 reflection

**Files:**
- Modify: `server/internal/bootstrap/server.go`

- [ ] **Step 1: 将无条件调用改为条件注册**

将第 60 行:
```go
reflection.Register(grpcServer)
```

替换为:
```go
if cfg.GRPC_REFLECTION == "true" {
    reflection.Register(grpcServer)
}
```

- [ ] **Step 2: 验证编译**

```bash
cd server && go build ./internal/bootstrap/...
```

Expected: 编译通过，无错误。

- [ ] **Step 3: 验证整包编译**

```bash
cd server && go build ./...
```

Expected: 全量编译通过，无错误。

---

### Task 3: .env.example — 文档化新环境变量

**Files:**
- Modify: `server/.env.example`

- [ ] **Step 1: 在 # Server 段落末尾新增注释**

在 `CORS_ALLOWED_ORIGINS=` 行之后（第 29 行之后），插入:

```text

# gRPC Reflection: set to "true" to enable gRPC server reflection.
# Disabled by default. Use grpcurl or grpcui locally for debugging.
# GRPC_REFLECTION=true
```

- [ ] **Step 2: 验证文件格式**

```bash
grep -n "GRPC_REFLECTION" server/.env.example
```

Expected: 匹配到新增行。

---

### Task 4: 验证 — 编译 + 现有测试

**Files:**
- (无变更，仅执行验证)

- [ ] **Step 1: 确认未开启 reflection 时编译通过**

```bash
cd server && go build ./...
```

Expected: 全量编译通过。

- [ ] **Step 2: 运行 platform 相关测试**

```bash
cd server && go test ./internal/config/... -v -count=1
```

Expected: 现有测试全部通过。

- [ ] **Step 3: 运行全量 Go vet**

```bash
cd server && go vet ./...
```

Expected: 无新增 vet 警告。

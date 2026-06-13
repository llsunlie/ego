# CORS 白名单设计

**日期**: 2026-06-13
**状态**: 已确认

## 背景

当前 CORS 配置允许任意来源访问，存在安全隐患：
- `server.go:58` — `WithOriginFunc` 无条件返回 `true`
- `server.go:97` — 回退 handler 设 `Access-Control-Allow-Origin: *`

需改为白名单机制，通过环境变量配置。

## 变更范围

仅涉及服务端，共 3 个文件：

| 文件 | 变更 |
|------|------|
| `server/internal/config/config.go` | 新增 `CORSAllowedOrigins` 字段 |
| `server/internal/bootstrap/server.go` | 替换两处通配符为白名单校验 |
| `server/.env.example` | 新增配置项说明 |

## 详细设计

### config.go

```go
CORSAllowedOrigins string // 从 CORS_ALLOWED_ORIGINS 读取，逗号分隔
```

新增方法：

```go
func (c *Config) AllowedOrigins() []string {
    if c.CORSAllowedOrigins == "" {
        return nil
    }
    origins := strings.Split(c.CORSAllowedOrigins, ",")
    for i := range origins {
        origins[i] = strings.TrimSpace(origins[i])
    }
    return origins
}
```

### server.go — origin 校验逻辑

```go
// isOriginAllowed 检查 origin 是否在白名单中。
// 白名单为空 → 拒绝所有（deny-by-default）。
// TLS_DOMAIN 为空（非生产环境） → 自动放行 localhost/127.0.0.1 任意端口。
func isOriginAllowed(origin string, allowed []string, tlsDomain string) bool {
    if origin == "" {
        return true // 同源请求无 Origin 头，放行
    }
    // 本地开发放行
    if tlsDomain == "" && isLocalhost(origin) {
        return true
    }
    for _, a := range allowed {
        if a == origin {
            return true
        }
    }
    return false
}
```

两处修复：

1. **gRPC-web `WithOriginFunc`**（line 58）— 替换为 `isOriginAllowed`
2. **回退 handler `Access-Control-Allow-Origin`**（line 97）— 替换 `"*"` 为动态匹配

### .env.example

```bash
# CORS: 允许的 Web 访问来源，逗号分隔
CORS_ALLOWED_ORIGINS=https://ego.app,https://www.ego.app
```

## 行为表

| 场景 | TLS_DOMAIN | CORS_ALLOWED_ORIGINS | Origin | 结果 |
|------|-----------|---------------------|--------|------|
| 同源请求 | 任意 | 任意 | 空 | 放行 |
| 本地开发 | 空 | 任意 | `http://localhost:*` | 放行 |
| 本地开发 | 空 | 任意 | `http://127.0.0.1:*` | 放行 |
| 生产，白名单内 | `ego.app` | `https://ego.app` | `https://ego.app` | 放行 |
| 生产，白名单外 | `ego.app` | `https://ego.app` | `https://evil.com` | 拒绝 |
| 生产，空白名单 | `ego.app` | 空 | `https://ego.app` | 拒绝 |

## 非目标

- 不修改 proto 契约
- 不修改客户端代码
- gRPC-web 库已处理 OPTIONS 预检，无需额外添加

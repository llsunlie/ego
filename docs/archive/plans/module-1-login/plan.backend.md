# Module 1 Backend — Login + gRPC 骨架 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建 Go gRPC 后端，实现 Login RPC（JWT + bcrypt 自动注册），连接 PostgreSQL。

**Architecture:** 单 Go module `server/`，gRPC server 监听 :50051，pgxpool 连接 PostgreSQL，JWT 通过 gRPC unary interceptor 验证，Login RPC 白名单跳过认证。

**Tech Stack:** Go 1.22+, protoc, gRPC, pgxpool, golang-jwt, bcrypt, google/uuid

---

### Task 1: Go module 初始化 + Makefile + 入口骨架

**Files:**
- Create: `server/go.mod`
- Create: `server/Makefile`
- Create: `server/cmd/ego/main.go`
- Create: `server/internal/config/config.go`

- [ ] **Step 1: 初始化 Go module**

```bash
cd server && go mod init ego-server
```

- [ ] **Step 2: 编写 Makefile**

```makefile
# server/Makefile
.PHONY: proto-gen migrate run

proto-gen:
	protoc --go_out=proto/ego --go_opt=paths=source_relative \
	       --go-grpc_out=proto/ego --go-grpc_opt=paths=source_relative \
	       ../proto/ego/api.proto

migrate:
	go run cmd/migrate/main.go

run:
	go run cmd/ego/main.go
```

- [ ] **Step 3: 编写 config.go**

```go
// server/internal/config/config.go
package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/ego?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		Port:        getEnv("PORT", "50051"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: 编写入口 main.go（骨架，暂无 gRPC 注册）**

```go
// server/cmd/ego/main.go
package main

import (
	"log"
	"net"

	"ego-server/internal/config"
	"ego-server/internal/db"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	server := grpc.NewServer()
	// TODO: 后续 task 注册 EgoServer

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("gRPC server listening on :%s", cfg.Port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
```

- [ ] **Step 5: 安装依赖**

```bash
cd server && go mod tidy
```

- [ ] **Step 6: 验证编译通过**

```bash
cd server && go build ./...
```
Expected: 编译成功（gRPC server 启动会 panic 因为没有注册服务，这是正常的）

---

### Task 2: Proto 定义 + 编译生成 Go 代码

**Files:**
- Create: `proto/ego/api.proto`
- Create: `server/proto/ego/` (generated, gitignore)

- [ ] **Step 1: 编写 proto 文件**

```protobuf
// proto/ego/api.proto
syntax = "proto3";

package ego;

option go_package = "ego-server/proto/ego";

// --- Service ---

service Ego {
  rpc Login(LoginReq) returns (LoginRes);
}

// --- Login ---

message LoginReq {
  string account  = 1;
  string password = 2;
}

message LoginRes {
  string token   = 1;
  bool   created  = 2;
}
```

- [ ] **Step 2: 安装 protoc 插件**

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

- [ ] **Step 3: 生成 Go 代码**

```bash
cd server && make proto-gen
```
Expected: 生成 `server/proto/ego/api.pb.go` 和 `server/proto/ego/api_grpc.pb.go`

- [ ] **Step 4: 验证生成代码可编译**

```bash
cd server && go build ./...
```
Expected: 编译成功

---

### Task 3: PostgreSQL 连接 + users migration

**Files:**
- Create: `server/internal/db/postgres.go`
- Create: `server/internal/db/migrations/001_users.sql`
- Create: `server/cmd/migrate/main.go`

- [ ] **Step 1: 编写 pgxpool 连接**

```go
// server/internal/db/postgres.go
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}
```

- [ ] **Step 2: 编写 users migration SQL**

```sql
-- server/internal/db/migrations/001_users.sql
CREATE TABLE users (
  id            UUID PRIMARY KEY,
  account       VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX idx_users_account ON users(account);
```

- [ ] **Step 3: 编写 migration 执行入口**

```go
// server/cmd/migrate/main.go
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"ego-server/internal/config"
	"ego-server/internal/db"
)

func main() {
	cfg := config.Load()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	sql, err := os.ReadFile(filepath.Join("internal", "db", "migrations", "001_users.sql"))
	if err != nil {
		log.Fatalf("read migration: %v", err)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migration 001_users applied")
}
```

> **注意：** 实际实现时应使用 `embed` 将 migration 文件嵌入二进制，或使用 golang-migrate 等工具。这里简化为直接读取文件用于开发阶段。

- [ ] **Step 4: 安装 pgx 依赖**

```bash
cd server && go get github.com/jackc/pgx/v5 && go mod tidy
```

- [ ] **Step 5: 验证 migration 可执行（需要 PostgreSQL 运行）**

```bash
cd server && go run cmd/migrate/main.go
```
Expected: `migration 001_users applied`

---

### Task 4: Auth 包 — JWT + gRPC Interceptor

**Files:**
- Create: `server/internal/auth/jwt.go`
- Create: `server/internal/auth/interceptor.go`

- [ ] **Step 1: 编写 JWT 签发与解析**

```go
// server/internal/auth/jwt.go
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseJWT(tokenStr string, secret []byte) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}
	return userID, nil
}
```

- [ ] **Step 2: 编写 gRPC unary interceptor**

```go
// server/internal/auth/interceptor.go
package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(jwtSecret []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Login RPC 白名单，跳过认证
		if strings.Contains(info.FullMethod, "Login") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}

		tokenStr := strings.TrimPrefix(values[0], "Bearer ")
		if tokenStr == values[0] {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		userID, err := ParseJWT(tokenStr, jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "user_id", userID)
		return handler(ctx, req)
	}
}
```

- [ ] **Step 3: 安装 jwt 依赖**

```bash
cd server && go get github.com/golang-jwt/jwt/v5 && go mod tidy
```

- [ ] **Step 4: 验证编译**

```bash
cd server && go build ./...
```
Expected: 编译成功

---

### Task 5: Login Handler

**Files:**
- Create: `server/internal/login/handler.go`
- Modify: `server/cmd/ego/main.go` (注册 EgoServer)

- [ ] **Step 1: 编写 Login handler**

```go
// server/internal/login/handler.go
package login

import (
	"context"
	"time"

	"ego-server/internal/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	db     *pgxpool.Pool
	jwtKey []byte
}

func NewHandler(db *pgxpool.Pool, jwtKey []byte) *Handler {
	return &Handler{db: db, jwtKey: jwtKey}
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	// 1. 查 users 表
	var userID, passwordHash string
	err := h.db.QueryRow(
		ctx,
		"SELECT id, password_hash FROM users WHERE account = $1",
		req.Account,
	).Scan(&userID, &passwordHash)

	if err != nil {
		// 用户不存在 → 自动注册
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash password")
		}

		userID = uuid.New().String()
		now := time.Now()

		_, err = h.db.Exec(
			ctx,
			"INSERT INTO users (id, account, password_hash, created_at) VALUES ($1, $2, $3, $4)",
			userID, req.Account, string(hash), now,
		)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create user")
		}

		token, err := auth.GenerateJWT(userID, h.jwtKey)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to generate token")
		}

		return &pb.LoginRes{Token: token, Created: true}, nil
	}

	// 用户存在 → bcrypt 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "密码错误")
	}

	token, err := auth.GenerateJWT(userID, h.jwtKey)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginRes{Token: token, Created: false}, nil
}
```

- [ ] **Step 2: 更新 main.go 注册 EgoServer**

```go
// server/cmd/ego/main.go (完整版)
package main

import (
	"log"
	"net"

	"ego-server/internal/auth"
	"ego-server/internal/config"
	"ego-server/internal/db"
	"ego-server/internal/login"

	"google.golang.org/grpc"

	pb "ego-server/proto/ego"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	jwtKey := []byte(cfg.JWTSecret)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(jwtKey)),
	)

	loginHandler := login.NewHandler(pool, jwtKey)
	pb.RegisterEgoServer(server, loginHandler)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("gRPC server listening on :%s", cfg.Port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
```

- [ ] **Step 3: 安装依赖**

```bash
cd server && go get github.com/google/uuid golang.org/x/crypto && go mod tidy
```

- [ ] **Step 4: 验证编译**

```bash
cd server && go build ./...
```
Expected: 编译成功

---

### Task 6: 端到端验证

- [ ] **Step 1: 启动 PostgreSQL（如未运行）**

```bash
# 确保 PostgreSQL 运行，创建 ego 数据库
createdb ego
```

- [ ] **Step 2: 执行 migration**

```bash
cd server && make migrate
```
Expected: `migration 001_users applied`

- [ ] **Step 3: 启动 gRPC server**

```bash
cd server && make run
```
Expected: `gRPC server listening on :50051`

- [ ] **Step 4: 使用 grpcurl 测试 Login（新用户自动注册）**

```bash
grpcurl -plaintext \
  -d '{"account": "test", "password": "123456"}' \
  localhost:50051 ego.Ego/Login
```
Expected: 返回 `{"token": "<jwt>", "created": true}`

- [ ] **Step 5: 再次 Login（已有用户，正确密码）**

```bash
grpcurl -plaintext \
  -d '{"account": "test", "password": "123456"}' \
  localhost:50051 ego.Ego/Login
```
Expected: 返回 `{"token": "<jwt>", "created": false}`

- [ ] **Step 6: 测试错误密码**

```bash
grpcurl -plaintext \
  -d '{"account": "test", "password": "wrong"}' \
  localhost:50051 ego.Ego/Login
```
Expected: 返回 `UNAUTHENTICATED` 错误，"密码错误"

---

## 验收标准

- [ ] `make run` 启动后端，gRPC 服务监听 :50051
- [ ] 新 account + password → Login 返回 `created: true` + JWT token
- [ ] 已有 account + 正确密码 → Login 返回 `created: false` + JWT token
- [ ] 错误密码 → 返回 UNAUTHENTICATED
- [ ] users 表中正确存储 bcrypt 哈希（非明文）
- [ ] JWT token 包含 user_id，有效期 30 天

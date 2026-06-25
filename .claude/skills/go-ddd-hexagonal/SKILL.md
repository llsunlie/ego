---
name: go-ddd-hexagonal
description: Use when starting a new Go backend project, designing server architecture, or when the user asks to follow the "ego project architecture" — a modular monolith with DDD bounded contexts, hexagonal/ports-and-adapters architecture, gRPC, and PostgreSQL
---

# Go DDD Hexagonal Architecture

## Overview

**Go modular monolith + DDD bounded contexts + hexagonal architecture (ports & adapters).** Each business domain is an independent module with its own `domain/app/adapter` layers. Infrastructure lives in `platform/`. Modules communicate only through contracts, read models, and domain events — never by importing each other's internals.

## Core Principle

```
Proto DTO → adapter/grpc → app (use cases) → domain (entities + ports) ← adapter/postgres
                                                                    ← adapter/ai
                                                                    ← adapter/eventbus
```

**Dependencies always point inward.** domain knows nothing about gRPC, SQL, AI, or any external SDK.

## When to Use

- Starting a new Go backend with multiple business domains
- Building API servers with gRPC + gRPC-web
- Projects that need clear module boundaries and independent evolvability
- When AI-generated insights, embeddings, or LLM calls are part of the system

**Do NOT use for:**
- Single-table CRUD apps — overkill
- Serverless functions — this is a long-running server pattern
- Microservices — this is a modular monolith, not distributed

## Directory Structure

```
project/
├── proto/                        # API contract source (.proto files)
│   └── {service}/
│       └── api.proto
│
├── server/
│   ├── cmd/
│   │   ├── {app}/main.go         # Server entrypoint (~30 lines)
│   │   └── migrate/main.go       # DB migration runner
│   ├── proto/                    # Generated Go code (protoc output, DO NOT EDIT)
│   │   └── {service}/
│   │       ├── api.pb.go
│   │       └── api_grpc.pb.go
│   ├── internal/
│   │   ├── config/               # Env-based config loading
│   │   ├── bootstrap/            # Composition root (wires everything)
│   │   ├── shared/               # Cross-cutting primitives only
│   │   │   ├── domain/           #   Common errors, event interface, clock
│   │   │   └── app/              #   Transaction runner, app-level helpers
│   │   ├── platform/             # Infrastructure (no business logic)
│   │   │   ├── auth/             #   JWT, bcrypt, gRPC interceptors
│   │   │   ├── postgres/         #   DB pool, sqlc generated code, migrations
│   │   │   ├── ai/               #   LLM client (chat + embedding)
│   │   │   ├── eventbus/         #   Domain event pub/sub
│   │   │   ├── logging/          #   Structured slog+zap logger
│   │   │   ├── metrics/          #   Prometheus metrics
│   │   │   └── ratelimit/        #   Token bucket rate limiter
│   │   └── {module}/             # One directory per bounded context
│   │       ├── domain/           #   Entities, value objects, repository interfaces, errors
│   │       ├── app/              #   Use cases, port interfaces, business orchestration
│   │       ├── adapter/
│   │       │   ├── grpc/         #   gRPC handler + proto↔domain mappers
│   │       │   ├── postgres/     #   Repository implementations
│   │       │   ├── ai/           #   AI prompt engineering (if module uses AI)
│   │       │   └── id/           #   ID generation (UUID)
│   │       └── module.go         #   Module DI: NewHandler(deps) → Handler
```

## Layer Responsibilities

| Layer | Contains | Forbidden |
|-------|---------|-----------|
| **domain/** | Entity, Value Object, Aggregate, Repository interface, Domain Event, business errors | proto, pgx, sqlc, grpc/status, config, any SDK |
| **app/** | Use cases, transaction boundaries, port interfaces, business orchestration | SQL strings, concrete AI SDK, proto types |
| **adapter/grpc/** | gRPC handler, proto↔domain mapper, request validation | Domain rules, business logic |
| **adapter/postgres/** | Repository impl, read model, sqlc/pgx adaptation, persistence model conversion | Business decisions, domain rules |
| **platform/** | Pure technical infrastructure: auth, DB, AI client, logging, metrics | Any business logic, any business RPC |
| **shared/** | Common ID wrappers, domain event interface, clock, transaction interface | Business entities, business rules |
| **bootstrap/** | Process-level DI, gRPC server lifecycle, module wiring | Business logic, SQL, AI prompts, business rules |

## The Hard Rules

These rules are what make the architecture work. The baseline agent naturally produces a reasonable structure but misses these constraints:

### 1. One Writer Per Table
Every business table has exactly one module that owns writes. Other modules may read through the owning module's contract, but never write directly.

```
users         → only identity writes
moments       → only writing writes
insights      → starmap writes constellation-level; writing writes moment-level
stars         → only starmap writes
chat_sessions → only conversation writes
```

Document table ownership in each module's `app/ports.go` as the authoritative interface.

### 2. Proto DTO Stays at the Boundary
Proto types (`pb.LoginReq`, `pb.CreateMomentRes`, etc.) are used **only** in `adapter/grpc/` and `bootstrap/`. They never enter `app/` or `domain/`. The gRPC handler maps proto→domain on the way in, and domain→proto on the way out.

### 3. Domain Models Carry No Framework Tags
Domain structs have no `json`, `db`, `protobuf` tags. Persistence models (in adapter/postgres) have sqlc/db tags. The adapter converts between them at the repository boundary.

### 4. Modules Never Import Each Other's Internals
Cross-module access goes through:
- **Go interface** defined in the providing module's `app/ports.go`
- **Read model struct** returned by the providing module's public query methods
- **Domain event** published to `platform/eventbus` for async, non-blocking coordination

Never: `import "project/internal/other-module/adapter/postgres"` or `import "project/internal/other-module/domain"`.

### 5. Read-Only Modules Exist
Some modules only query data owned by others (e.g., timeline reads moments but never writes). They have no write repository — only read models. This is lightweight CQRS within a monolith.

## Two-Level Dependency Injection

### Level 1: Process (`bootstrap/platform.go`)
```go
type Platform struct {
    Pool     *pgxpool.Pool
    JWTKey   []byte
    Hasher   PasswordHasher
    Tokens   TokenIssuer
    Logger   *slog.Logger
    AIClient *ai.Client
    EventBus *eventbus.Bus
}

func InitPlatform(cfg *config.Config) (*Platform, error) {
    // Create DB pool, logger, JWT, AI client, event bus
    // These are process-level singletons
}
```

### Level 2: Module (`internal/{module}/module.go`)
```go
type Deps struct {
    DB        sqlc.DBTX        // the DBTX interface, not *pgxpool.Pool
    Hasher    app.PasswordHasher
    Tokens    app.TokenIssuer
    AIClient  *ai.Client       // only if module needs AI
}

func NewHandler(deps Deps) *grpc.Handler {
    queries := sqlc.New(deps.DB)
    repo := postgres.NewRepo(queries)
    ids := id.NewUUIDGenerator()
    useCase := app.NewUseCase(repo, ids)
    return grpc.NewHandler(useCase)
}
```

**Critical**: Module `Deps` depends on **interfaces** (often from its own `app/ports.go`), not on concrete platform types (except `*ai.Client`). The bootstrap layer adapts platform implementations to module interfaces.

## Composite Handler Pattern

gRPC allows only one service registration. All module handlers are composed into a single struct that routes each RPC to its owning module:

```go
// bootstrap/composite.go
type AppHandler struct {
    pb.UnimplementedAppServer
    identity pb.AppServer  // only implements auth RPCs
    writing  pb.AppServer  // only implements moment RPCs
    timeline pb.AppServer  // only implements query RPCs
}

func (h *AppHandler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
    return h.identity.Login(ctx, req)
}

func (h *AppHandler) CreateMoment(ctx context.Context, req *pb.CreateMomentReq) (*pb.CreateMomentRes, error) {
    return h.writing.CreateMoment(ctx, req)
}
```

## Cross-Module Communication

| Pattern | When | Example |
|---------|------|---------|
| **Direct interface call** | Synchronous, same process | writing calls identity to verify user exists |
| **Read model query** | One module reads another's data | timeline queries moments through writing's public reader |
| **Domain event** | Async, fire-and-forget | "MomentCreated" → starmap checks for constellation clustering |
| **app/ports.go interface** | Module's public API contract | Other modules depend on the interface, not the implementation |

## Technology Stack Conventions

| Component | Choice | Notes |
|-----------|--------|-------|
| Language | Go 1.22+ | |
| API | gRPC + gRPC-web | `improbable-eng/grpc-web` for browser compatibility |
| DB | PostgreSQL + pgvector | Embedding similarity search |
| SQL gen | sqlc | Type-safe, no ORM magic |
| Auth | JWT (access + refresh) | bcrypt for password hashing |
| AI | OpenAI-compatible API | Separate endpoints for chat vs embedding |
| Logging | slog + zap | Structured, context-propagated |
| Metrics | Prometheus | HTTP + gRPC metrics |
| TLS | autocert (Let's Encrypt) | Automatic cert management |

## How to Apply to a New Project

When the user says "按照这个架构来" or "follow the ego architecture":

1. **Identify bounded contexts** — list the business domains (Auth, Writing, Browsing, Insights, etc.)
2. **Define table ownership** — one writer per table, documented in each module's `app/ports.go`
3. **Scaffold the directory** — `server/` with `cmd/`, `proto/`, `internal/{config,bootstrap,shared,platform,{module1},{module2}...}`
4. **Design the proto file** — all RPCs in one `.proto`, gRPC service per context group
5. **Build platform first** — DB pool, auth, logging, AI client
6. **Implement modules one at a time** — domain → app → adapter → module.go
7. **Wire in bootstrap** — `main.go` → `InitPlatform` → each `NewHandler` → composite → `NewServer`

## Common Mistakes

| Mistake | Reality |
|---------|---------|
| "I'll put DB queries in app layer" | app layer has no SQL. Put queries in adapter/postgres. |
| "Just import the other module's repo directly" | Cross-module access goes through `app/ports.go` interfaces, never internal imports. |
| "Proto types are my domain model" | Proto DTOs stay in adapter/grpc. Domain has its own types. |
| "Everything in one big module" | Separate by who writes the data. Read-only contexts are valid modules. |
| "Platform can contain business helpers" | Platform has zero business logic. Business strategies go in the owning module's app/. |
| "I'll define module boundaries later" | Boundaries erode fast. Define `app/ports.go` interfaces first — they are the contract. |

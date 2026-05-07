# identity Architecture

Bounded context: Identity.

Answers the single question: **who is making this request?**

## Responsibilities

- Account + password login
- Automatic registration when account does not exist
- bcrypt password hashing and verification
- JWT token issuance orchestration

## Primary use cases

| Use Case | Description |
| --- | --- |
| Login | Find user by account; if not found → auto-register; if found → verify password; issue JWT |

## Internal structure

```
internal/identity/
├── domain/
│   ├── user.go              # User entity + UserRepository interface
│   └── errors.go            # ErrUserNotFound
├── app/
│   ├── login.go             # LoginUseCase: orchestrate login/register flow
│   └── ports.go             # PasswordHasher, TokenIssuer interfaces
├── adapter/
│   ├── grpc/
│   │   └── handler.go       # Thin handler: proto → app call → gRPC response
│   └── postgres/
│       └── user_repo.go     # UserRepository: sqlc Queries → domain.User mapping
├── AGENTS.md
├── ARCHITECTURE.md
├── CONTRACT.md
└── .harness/
```

## Layer rules

| Layer | Contains | Must NOT contain |
| --- | --- | --- |
| `domain/` | User entity, UserRepository interface, domain errors | proto, pgx, sqlc, bcrypt, JWT, gRPC status |
| `app/` | LoginUseCase, PasswordHasher port, TokenIssuer port, LoginResult DTO | SQL strings, concrete AI/DB clients, proto types |
| `adapter/grpc/` | Handler, proto ↔ app mapping, error → gRPC status mapping | business rules, password hashing, DB access |
| `adapter/postgres/` | UserRepository impl, sqlc → domain mapping | business decisions, password logic |

## Domain model

```
User
├── ID           string      (uuid)
├── Account      string
├── PasswordHash string      (bcrypt)
└── CreatedAt    time.Time
```

## Port interfaces (defined in app/)

- **PasswordHasher**: `Hash(plaintext) → (hash, error)` / `Verify(hash, plaintext) → error`
- **TokenIssuer**: `Issue(userID) → (token, error)`

These are implemented by `platform/auth` and wired in `bootstrap/`. Identity's app layer does not import platform; platform does not import identity — Go implicit interface satisfaction makes this work.

## Dependency direction

```
adapter/grpc ──▶ app ──▶ domain
                     │
adapter/postgres ────┘
      │
      ▼
platform/postgres (sqlc)

domain ← no external deps
```

## Wiring (in bootstrap/)

1. Create `sqlc.Queries` from `pgxpool.Pool`
2. Create `postgres.UserRepository` (wraps queries)
3. Create `BcryptHasher` (from platform/auth)
4. Create `JWTIssuer` (from platform/auth)
5. Create `LoginUseCase(repo, hasher, issuer)`
6. Create `grpc.Handler(useCase)`

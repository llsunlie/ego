# identity Architecture

Bounded context: Identity.

Answers the single question: **who is making this request?**

## Responsibilities

- Account + password login
- Automatic registration when account does not exist
- bcrypt password hashing and verification
- JWT token issuance orchestration

## Layer design

| Layer | Role |
| --- | --- |
| `domain/` | User aggregate + UserRepository port + domain errors |
| `app/` | LoginUseCase + ports: PasswordHasher, TokenIssuer, IDGenerator |
| `adapter/grpc/` | Thin handler: proto → app call → gRPC response with error mapping |
| `adapter/postgres/` | UserRepository implementation via sqlc |
| `adapter/id/` | UUID generator |
| `module.go` | Module-level composition |

## Dependency flow

```
bootstrap/identity.go → identity.Deps{DB, Hasher, Tokens} → identity/module.go
  → postgres UserRepo
  → adapter/id UUIDGenerator
  → app LoginUseCase
  → gRPC handler
```

Hasher and Tokens are infrastructure capabilities provided by bootstrap (from platform/auth).

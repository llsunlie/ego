# identity Architecture

Bounded context: Identity.

Responsibilities:

- Login
- Automatic registration
- User lookup
- Password verification
- Token issuance orchestration

Current implementation is still thin and lives in `adapter/grpc/handler.go`; future work should move business orchestration into `app/` and user concepts into `domain/`.


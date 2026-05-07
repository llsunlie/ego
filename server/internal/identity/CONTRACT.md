# identity Contract

## Owned writes

- `users`

Identity has exclusive write access to the `users` table. No other module may insert, update, or delete user records.

## Exposed to other modules

### Via gRPC auth interceptor

All non-Login RPCs receive a verified `user_id` injected into the request context by the auth interceptor. Other modules call `ctx.Value("user_id")` to obtain the current user identity.

### RPC

| RPC | Input | Output | Notes |
| --- | --- | --- | --- |
| `Login` | `account` + `password` | `token` (JWT) + `created` (bool) | Auto-registers when account is new |

## Reads from other modules

None. Identity is self-contained and does not depend on any other business module.

## Constraints for other modules

- Must NOT query `users.password_hash` directly
- Must NOT insert or update `users` rows
- Must NOT call `platform/auth` JWT functions directly — authentication is handled by the interceptor
- Must obtain user identity only via `ctx.Value("user_id")` from the gRPC auth interceptor

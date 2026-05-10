# identity Contract

## Owned writes

- `users`

## RPCs

| RPC | Input | Output | Notes |
| --- | --- | --- | --- |
| `Login` | `account` + `password` | `token` (JWT) + `created` (bool) | Auto-registers when account is new |

## Provided capabilities

- JWT parsing utilities in `platform/auth`
- Bcrypt hashing through `platform/auth.BcryptHasher`

## Reads from other modules

None. Identity is self-contained.

## Constraints for other modules

- Must NOT query `users.password_hash` directly
- Must NOT insert or update `users` rows
- Must obtain user identity only via `ctx.Value("user_id")` from the gRPC auth interceptor

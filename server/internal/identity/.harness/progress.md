# identity Progress

## Current State

- Login handler and tests have been moved into the identity bounded context.
- Business logic is still colocated in the gRPC adapter and should be split into `app` later.

## Next Best Step

- Extract Login use case into `identity/app` when identity receives new behavior.


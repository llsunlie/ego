# identity Agent Guide

`identity` owns user identity, Login, and automatic registration.

## Rules

- Owns writes to `users`.
- May use platform auth primitives for JWT and password hashing.
- Must not read or write Moment, Star, Constellation, or Chat data.
- gRPC mapping belongs in `adapter/grpc`.


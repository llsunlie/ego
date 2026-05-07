# identity Contract

Identity currently exposes the `Login` RPC through `adapter/grpc`.

Owned table:

- `users`

Other modules receive authenticated `user_id` from the gRPC auth interceptor. They should not query password hashes or identity internals.


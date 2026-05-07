# Proto Contract

This directory contains the source `.proto` files for ego's client-server API contract.

## Rules

- `proto/ego/api.proto` is the only hand-edited gRPC contract source.
- Frontend and backend generated files must be regenerated from this source.
- Do not edit generated files as the source of truth.
- Additive changes are preferred. Do not reuse removed field tags.
- Breaking changes must be recorded in `CHANGELOG.md`.

## Generated Outputs

Current expected generated locations:

```text
server/proto/ego/*.pb.go      Go server/client generated code
client/...                    Frontend generated code, location defined by client tooling
```

`server/proto/` is intentionally kept as generated Go output, not as the canonical contract source.


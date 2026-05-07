# Server Proto Generated Code

This directory contains Go code generated from the root API contract:

```text
../../proto/ego/api.proto
```

## Rules

- Do not hand-edit `*.pb.go` or `*_grpc.pb.go`.
- Regenerate from `server/` with:

```text
make proto-gen
```

- If the API contract itself needs to change, edit `../../proto/ego/api.proto` first.
- Record contract changes in `../../proto/CHANGELOG.md`.


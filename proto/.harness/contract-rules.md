# Proto Contract Rules

`proto/ego/api.proto` is the canonical API contract between frontend and backend.

## Rules

- Only edit `.proto` source files under root `proto/`.
- Do not hand-edit generated files under `server/proto/`.
- Field tags are permanent. Never reuse a deleted tag.
- Prefer adding fields over changing existing field meaning.
- Record breaking changes in `proto/CHANGELOG.md`.
- After changing proto source, regenerate backend and frontend outputs.

## Backend Generation

From `server/`:

```text
make proto-gen
```

This reads `../proto/ego/api.proto` and writes Go generated files into `server/proto/ego/`.


# Server Clean State Checklist

- Run or intentionally skip `go test ./...` with a recorded reason.
- Run or intentionally skip `go build ./...` with a recorded reason.
- Update the current module's `.harness/progress.md`.
- Update the current module's `.harness/feature_list.json` when feature state changes.
- Do not leave temporary debug code.
- Do not modify another module owner's implementation without recording the reason.
- Update proto harness files when API contracts change.


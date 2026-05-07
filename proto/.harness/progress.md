# Proto Progress

## Current State

- Root `proto/ego/api.proto` is the canonical source contract.
- `server/proto/ego/*.pb.go` is generated Go output.
- Backend generated output still needs regeneration in an environment with `make`/`protoc` available after proto changes.

## Next Best Step

- Regenerate backend proto output whenever `proto/ego/api.proto` changes.

## Verification Notes

- `make proto-gen` could not run in the current PowerShell environment because `make` is not installed.
- The direct `protoc` command also could not run because `protoc` is not installed.

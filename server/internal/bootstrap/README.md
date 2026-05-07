# bootstrap

Dependency assembly for the server process. Wires together config, platform adapters, application use cases, and gRPC handlers.

Business logic does not belong here.

## Structure

```
internal/bootstrap/
├── platform.go     # InitPlatform: DB pool, JWT, hasher, token issuer
├── identity.go     # NewIdentityHandler: UserRepo → LoginUseCase → Handler
├── server.go       # NewServer: gRPC + gRPC-web lifecycle
└── README.md
```

## Usage in main.go

```go
cfg := config.Load()
p, _ := bootstrap.InitPlatform(cfg)
defer p.Close()
identityHandler := bootstrap.NewIdentityHandler(p)
server := bootstrap.NewServer(cfg, p, identityHandler)
server.Serve()
```

## Adding a new module

When adding a new bounded context (e.g. writing):

1. Create `bootstrap/writing.go` with `NewWritingHandler(p *Platform) pb.EgoServer`
2. Wire the module's dependencies inside that function
3. In `server.go`, pass the new handler alongside existing ones

Future modules will require a composite handler to satisfy `pb.RegisterEgoServer` since gRPC only allows one registration per service.

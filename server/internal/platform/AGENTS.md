# platform Agent Guide

`platform` owns backend infrastructure capabilities.

It may provide concrete implementations for interfaces requested by business modules, but it must not own ego business workflows.

## Allowed Responsibilities

- PostgreSQL pool, migrations, sqlc, transaction helpers
- JWT and password hashing primitives
- gRPC server plumbing and interceptors
- AI client adapters, prompt templates, output validators
- In-memory event bus or outbox infrastructure

## Forbidden Responsibilities

- Creating Moments
- Stashing Traces
- Forming Constellations as a business rule
- Managing ChatSessions
- Returning proto responses directly for business flows


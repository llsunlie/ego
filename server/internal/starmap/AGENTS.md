# starmap Agent Guide

`starmap` owns Trace stashing, Stars, Constellations, and constellation detail assets.

## Rules

- Owns writes to star and constellation-related tables.
- Reads Trace/Moment data through Writing contracts.
- Does not create Moments.
- Long-running clustering and AI asset generation should be event-driven or asynchronous.


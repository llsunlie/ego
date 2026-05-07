# writing Agent Guide

`writing` owns the "write now" flow: Trace, Moment, Echo, and current-session insight.

## Rules

- Owns writes to `moments` and future `traces`.
- Does not create Star or Constellation records.
- Does not manage chat sessions.
- AI usage must go through ports, not concrete SDK clients.


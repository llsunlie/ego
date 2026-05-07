# conversation Agent Guide

`conversation` owns chat sessions with past-self personas.

## Rules

- Owns writes to chat session and chat message tables.
- Reads PastSelfCard through Starmap contracts.
- Reads Moment context through Writing contracts.
- AI replies must preserve citation/reference constraints.


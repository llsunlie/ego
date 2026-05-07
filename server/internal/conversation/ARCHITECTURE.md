# conversation Architecture

Bounded context: Conversation.

Responsibilities:

- Start or restore ChatSession.
- Append user ChatMessage.
- Generate past-self reply through an interface.
- Validate citations and scope.
- Store AI replies and references.

Primary use cases:

- StartChat
- SendMessage


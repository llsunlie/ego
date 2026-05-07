# writing Contract

Owned writes:

- `moments`
- future `traces`

Expected read contracts for other modules:

- TraceReader
- MomentReader

Other modules may read Moments through explicit contracts, but must not create or update Moments directly.


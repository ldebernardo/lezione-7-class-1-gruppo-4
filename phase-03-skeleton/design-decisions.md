# Design Decisions - Warehouse Go Domain

## Article is the only aggregate root

Decision: callers mutate inventory only through `Article`.

Rule protected: `reserved <= quantity`, `quantity >= 0`, `price > 0`, name not
empty, and currency immutability all stay inside one consistency boundary.

Rejected alternative: an independent `InventoryRepository`.

Clean architecture alignment: the domain exposes business behavior, not storage
tables.

## SKU and Money are value objects

Decision: `SKU` and `Money` have no identity and compare by value.

Rule protected: invalid SKU, negative money and malformed currency are rejected
at construction time.

Rejected alternative: passing raw strings and integers through the aggregate.

Clean architecture alignment: validation belongs to the domain model, not HTTP
handlers or database constraints.

## Events are recorded, not published

Decision: `Article` appends domain events to a pending list and exposes a drain
method.

Rule protected: the domain has no messaging side effects.

Rejected alternative: publishing to a broker directly from the aggregate.

Clean architecture alignment: a later application/infrastructure layer can drain
and publish after persistence succeeds.

## Domain packages do not import infrastructure

Decision: `entities`, `events` and `interfaces` import no database driver or web
framework.

Rule protected: this phase remains domain-only.

Rejected alternative: adding MySQL persistence or `/articles` handlers now.

Clean architecture alignment: persistence and HTTP arrive in later checkpoints,
outside the domain.

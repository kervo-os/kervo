# kervo

Workspace context compiler. See `docs/` for PRD / RFC / ARCH.

```
make build && ./kervo
```

Structure follows ARCH-0001 (hexagonal): `internal/core` is pure and
must not import `internal/adapters` — enforced by `make arch-check`.

# Improvement plans

Audit against working tree on top of `33ecdbb` (in-flight uncommitted changes
present at audit time).

- **[001-fix-luks-detection.md](001-fix-luks-detection.md)** — urgent bug, full
  executor-ready plan. LUKS detection always returns false.
- **[PLAN.md](PLAN.md)** — backlog of the remaining findings (A–E) + direction
  notes. Promote an item to its own `00N-*.md` when you start it.

## Status

| # | Plan | Effort | Status |
|---|------|--------|--------|
| 001 | Fix LUKS detection | S | DONE (2026-06-25) |
| A | e2e tests skip (stale image names) | S | backlog → PLAN.md |
| B | Default log file litters cwd | S | backlog → PLAN.md |
| C | `http.Get` no timeout | S | backlog → PLAN.md |
| D | UEFI vars sized from CODE blob | S | backlog → PLAN.md |
| E | LUKS unit tests assert nothing | XS | fold into 001 |

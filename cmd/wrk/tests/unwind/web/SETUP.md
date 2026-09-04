# Scenario

**Feature**: `wrk --unwind --web` serves action preview + Run

```
wrk --unwind --web [--port] [--merge-back|--done] [--tag-next] [--push] …
  -> listen :port (auto from 8080; skip busy); stdout http://127.0.0.1:<port>/
  -> GET / HTML preview; POST /plan refreshes JobPlan; POST /run applies
  -> reject --show-graph / --verify / --dry-run / --dev
```

## Preconditions

- Inherits `cmd/wrk/tests/unwind` Request/Response/Run.
- Reject/help leaves are L2 InProcess (no long-running server).
- Handler/HTML/live-plan coverage is L1 in `wrkcli/unwind/web` (POST `/plan` full flags, remaining `apply_hint`; `#preview` is SVG action DAG from `epochs`).

## Steps

1. Grouping for `--unwind --web`.
2. Reject leaves fail at flag validation.
3. Help leaf runs `wrk -h`.

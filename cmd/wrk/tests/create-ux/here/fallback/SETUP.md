# Scenario

**Feature**: `--here` without shell integration falls back to nested bash at worktree

```
WRK_FOLLOWUP_FILE unset
wrk -t 'task' --open-in-agent --here
  -> warning: bash integration install hint
  -> nested bash at worktree (fake shell in tests)
```

## Preconditions

- Every successful fallback leaf **must** call `installFakeBashUX`.
- Follow-up channel stays closed (`UseFollowupEnv` false by default).

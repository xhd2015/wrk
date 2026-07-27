# Scenario

**Feature**: global `--no-config` skips reading `$WRK_HOME/config.json` on plain create

```
# plain create (no <target-dir>): applyConfig = (spawnTarget == "") && !noConfig
full create.* config + wrk --no-config
  -> native create only; config UX NOT applied

full create.* config + wrk --no-config --new-window|--new-terminal|--open-in-agent
  -> UX from CLI flags only (config still not base-merged)

corrupt / missing config.json + wrk --no-config
  -> never open/read config; no parse error; bare create succeeds
```

## Preconditions

- Git available; isolated `{WRK_HOME}`; `WRK_DATE=2026-06-30`.
- Same hermetic UX mocks as sibling create-ux groups (`installCreateUXMocks`).
- Plain create: no `SpawnDir` — worktree lands under `{WRK_HOME}/worktrees/…`.
- Contrast: `pipeline/config/defaults-match-flags` applies the same full config **without**
  `--no-config` and **does** drive space/iterm/agent.

## Steps

- Group Setup: seed main repo, install darwin UX mocks.
- Leaves write config (or corrupt payload) and set `req.Args` including `--no-config`.

## Context

- `--no-config` is long-only (no short alias); top-level main flag set.
- One-shot create UX CLI flags still parse and apply when present.
- Mutually exclusive with `--set-config` (covered under `set-config/mutual-exclusion/with-no-config`).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```

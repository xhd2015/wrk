# Scenario

**Feature**: `--no-config` is mutually exclusive with `--set-config`

```
wrk --no-config --set-config --show
  -> non-zero
  -> stderr: --no-config is mutually exclusive with --set-config
  -> no config.json write; no create side effects
```

## Steps

1. Start with empty WRK_HOME (no `config.json`).
2. Run `wrk --no-config --set-config --show` (any set-config subpath is rejected the same way).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	// --no-config before --set-config (order should not matter; pin common form).
	req.Args = append([]string{"--no-config"}, setConfigArgs("--show")...)
	return nil
}
```

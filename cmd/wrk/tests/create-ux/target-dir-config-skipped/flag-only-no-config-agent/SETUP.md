# Scenario

**Feature**: with `<target-dir>`, config agent does not merge when only terminal flag is set

```
# config: agent enabled (window off); no terminal in config
wrk <myrepo> <target-dir> --new-terminal
  -> iterm ForceNew at target
  -> NO agent follow-up (config agent skipped)
  -> NO outer agent-run
  -> NO space (no window flag; config window absent)
```

## Steps

1. Write create config with agent on only (terminal/window off).
2. Run with `--new-terminal` only (no `--open-in-agent`, no window flag).
3. Expect terminal from flag; agent stays off because config base is not applied.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeCreateUXConfig(t, req.WrkHome, map[string]interface{}{
		"agent": map[string]interface{}{
			"enabled":         true,
			"runner":          "grok-tty",
			"prompt_template": "/brainstorm ${task}",
			"args":            []string{"--session-id-from-prompt", "--no-submit", "--open"},
		},
	})
	req.Args = []string{"--new-terminal"}
	return nil
}
```

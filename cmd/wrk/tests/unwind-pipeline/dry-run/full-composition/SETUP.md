# Scenario

**Feature**: complete dry-run composition is accepted with rel peel display

```
# linked dep under external/dep; all unwind modifiers
linked dependency -> all unwind modifiers -> planned commit through reinstall
  -> would: peel external/dep … generate … reinstall
```

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.Args = []string{
		"--unwind",
		"--gen-commit-msg", "--commit", "--agent-runner=commandcode",
		"--merge-back", "--sync", "--tag-next", "--push",
		"--reinstall-local", "--dry-run",
	}
	return nil
}
```

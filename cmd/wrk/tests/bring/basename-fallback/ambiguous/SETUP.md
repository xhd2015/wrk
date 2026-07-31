# Scenario

**Feature**: multiple saved deps with same basename for --bring resolution (resolve once)

```
# two+ saved dep projects match basename
# preflight resolves each --bring arg once (TTY Select or non-TTY list error)
# apply reuses resolved abs path — no second prompt / no double dump
TTY + stdin -> one numbered select -> --bring succeeds for chosen dep
non-TTY -> error listing all candidate absolute paths once
```

## Steps

- Descendants seed two saved dep repos sharing basename `mydep` at different parent paths.
- Run `wrk --bring mydep` from consumer cwd without local `./mydep`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBringBasenameFallbackHelpersUsed()
	return nil
}
```

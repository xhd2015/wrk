# Scenario

**Feature**: tidy failure on one module is soft; other modules still pin; exit 0

```
root consumer requires dep (tidy ok)
bad/ requires dep + example.com/missing (tidy fails offline after pin)
  -> wrk --pin-locals
  -> pin both consumers for dep
  -> warning: go mod tidy in <bad-dir>: …
  -> root still has relative replace
  -> summary tidy failed >= 1
  -> exit 0
```

## Steps

1. Build tidy-fail-continues fixture (GOPROXY=off).
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTidyFailContinues(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```

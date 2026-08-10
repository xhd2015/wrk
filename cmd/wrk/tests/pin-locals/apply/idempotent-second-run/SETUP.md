# Scenario

**Feature**: second apply is idempotent (already up to date)

```
fixture needs pin
  -> wrk --pin-locals (first, in Setup)
  -> wrk --pin-locals (second, Run)
  -> second: already up to date / applied 0
  -> go.mod stable after first
```

## Steps

1. Build multi-repo external consumer fixture.
2. Run apply once in Setup; snapshot go.mod.
3. Run apply again via doctest Run.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMultiRepoExternalConsumer(t, req)
	req.Args = []string{"--pin-locals"}
	first := runPinLocalsOnce(t, req)
	if first.ExitCode != 0 {
		// Classic TDD: first run may also fail until product lands; still proceed
		// so second-run asserts stay meaningful once GREEN. Record baseline anyway.
		t.Logf("first pin-locals exit=%d stdout=%q stderr=%q", first.ExitCode, first.Stdout, first.Stderr)
	}
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	return nil
}
```

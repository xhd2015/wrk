# Scenario

**Feature**: with `--no-config`, corrupt `config.json` is never opened → no parse error

```
# invalid JSON on disk
{WRK_HOME}/config.json = "{not-json"
wrk --no-config
  -> exit 0; native create only
  -> no config parse/load error on stderr
  -> no space / iterm / agent
```

## Steps

1. Write a non-JSON payload to `{WRK_HOME}/config.json`.
2. Run bare create with `["--no-config"]`.
3. Expect success as if config were absent; no JSON/parse error.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Deliberately invalid JSON — without --no-config a load would fail.
	writeFile(t, filepath.Join(req.WrkHome, "config.json"), "{not-json\n")
	req.Args = []string{"--no-config"}
	return nil
}
```

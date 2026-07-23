# wrk --help composition docs — short-path help (L2 in-process CLI)

## Version
0.0.1

**Layer: L2 in-process CLI** — `wrkcli.RunCLI` with captured writers.
Not product-binary e2e (short path / usage only).

# DSN (Domain Specific Notion)

- **wrk CLI** — root or mode help documents flags on stdout; exit 0; no git required.
- **In-process** — same dispatch as production binary via `RunCLI`.

## How to Run

```sh
doctest test cmd/wrk/tests/done-compose/help
```

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string
	Args     []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var stdout, stderr bytes.Buffer
	code := wrkcli.RunCLI(req.Args, wrkcli.RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Dir:     req.RepoDir,
		WrkHome: req.WrkHome,
	})
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
	}, nil
}
```

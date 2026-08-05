# Scenario

**Feature**: non-full-URL PR refs are rejected

```
bare number / owner/repo#N / scheme-less path
  -> non-zero; error names full GitHub pull request URL
```

## Preconditions

- Positional present but not an accepted full URL.

## Steps

- Leaves set one invalid ref form each.

## Context

- Accepted only: `https://github.com/<owner>/<repo>/pull/<N>` (optional http).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```

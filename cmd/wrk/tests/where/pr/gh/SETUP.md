# Scenario

**Feature**: `--where --pr` requires working `gh` for PR view

```
missing gh on PATH  OR  gh pr view fails
  -> non-zero; empty stdout; clear stderr
```

## Preconditions

- Valid full URL argv; git fixtures may be present.

## Steps

- Leaves strip gh or force view failure.

## Context

- Same messaging style family as existing bare `--pr` missing-gh when possible.

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

# Scenario

**Feature**: `--where --pr` remains exclusive with other modes

```
--where --pr URL + --status (or other exclusive)
  -> non-zero; mutually exclusive / mode conflict
```

## Preconditions

- Compose allows only `--where` + `--pr` (+ URL). Still exclusive with
  `--status`, `--main`, `--list`, `--cd`, `--done`, etc.

## Steps

- Leaves add one forbidden partner flag.

## Context

- Regression guard so compose does not open the floodgates.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = req
	return nil
}
```

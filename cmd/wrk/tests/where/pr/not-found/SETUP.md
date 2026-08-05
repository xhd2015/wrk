# Scenario

**Feature**: valid URL + successful gh view, but no printable local path

```
gh ok + headRefName known
  -> no matching local project OR head not checked out
  -> non-zero; empty stdout; stderr names PR / head / repo
```

## Preconditions

- Fake gh returns headRefName successfully.

## Steps

- Leaves vary project match vs branch checkout absence.

## Context

- Distinct from invalid URL / missing gh (those are validation/gh groups).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = req
	return nil
}
```

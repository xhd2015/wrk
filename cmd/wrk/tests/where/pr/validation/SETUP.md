# Scenario

**Feature**: reject invalid / incomplete PR refs for `--where --pr` compose

```
missing URL / bare number / shorthand / scheme-less / --pr=URL / extra args
  -> non-zero before location lookup
  -> clear error; prefer “full GitHub pull request URL”
```

## Preconditions

- Compose mode selected (`--where` + `--pr`); ref is not a valid full URL form.

## Steps

- Leaves set invalid argv; fixtures minimal when needed.

## Context

- Full URL only (https or optionally http); host github.com.

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

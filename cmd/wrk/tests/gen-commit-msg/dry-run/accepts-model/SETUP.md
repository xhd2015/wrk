# Scenario

**Feature**: wrk --gen-commit-msg --dry-run --model accepts model without agent

```
# model flag is forwarded and accepted under dry-run
repo/ (1 staged) -> wrk --gen-commit-msg --dry-run --model some/model
  -> mock B success; exit 0
```

## Preconditions

- One staged text file.
- Explicit non-default model string `some/model`.

## Steps

1. Stage one text file.
2. Run with `--dry-run --model some/model`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFile(t, req)
	req.Args = []string{"--gen-commit-msg", "--dry-run", "--model", "some/model"}
	return nil
}
```

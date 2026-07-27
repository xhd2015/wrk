# Scenario

**Feature**: list-mode exclusives remain; multi-stage pipeline pairs without list are allowed under activeRoot

```
# list modes stay exclusive with pipeline stages
wrk --tag-next --list -> still mutually exclusive
wrk --reinstall-local --list -> still mutually exclusive

# multi-stage without done is allowed (classic RED unlock until implementer):
wrk --reinstall-local --sync -> NOT mutually exclusive
wrk --gen-commit-msg --sync -> NOT mutually exclusive
# full activeRoot coverage: compose-pipeline/
```

## Preconditions

- Standalone exclusives such as `--tag-next` + `--list` stay rejected.
- `--reinstall-local --list` stays exclusive with list mode.
- Bare multi-stage (`--gen-commit-msg --sync`, `--reinstall-local --sync`) is **allowed** under
  the activeRoot compose model (leaves under this group assert unlock; deeper e2e under
  `compose-pipeline/`).

## Steps

- Leaves set mode pairs (list exclusives stay RED-on-regression; pipeline pairs assert allow).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: regression exclusives need a valid git cwd for mode flags.
	skipIfNoGit(t)
	return nil
}
```

# Scenario

**Feature**: `wrk --merge-back --sync` runs sync from main after successful merge-back (no remove)

```
# primary --merge-back succeeds then runSync(main TargetPath); source wt stays
linked wtA (+ optional wtB behind)
  -> wrk --merge-back -y --sync
  -> merge without remove (message on stdout)
  -> blank line
  -> runSync(main): pass2 distribute / zero-summary
```

## Preconditions

- Git available; monotree root composition helpers.
- Composition not implemented yet → leaves **RED**.

## Steps

- Grouping only: leaves set fixture + `req.Args`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

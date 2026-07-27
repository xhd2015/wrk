# Scenario

**Feature**: current script with missing markers reports updated

```
pre-seed current embedded bash.sh only (no profile markers)
wrk --bash-integration --install
  -> bash integration: updated
  -> script (is up to date); both markers (marker installed)
```

## Steps

1. Seed `bash.sh` with the live embedded script content (`SeedCurrentScript`).
2. Leave profiles empty (no markers).
3. Run install.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "install")
	if req.DryRun {
		t.Fatalf("expected real install, not dry-run")
	}
	req.SeedCurrentScript = true
	return nil
}
```

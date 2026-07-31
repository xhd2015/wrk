# Scenario

**Feature**: `wrk -h` documents multi-mode `--pr` (show, status, comment, create, push)

```
workspace/ -> wrk -h
  -> exit 0
  -> help mentions --pr, --title, --comment
  -> --pr block documents with --status (PR metadata/checks)
  -> push-with-pr / open-PR tip-push rule is visible in --pr or --push help
```

## Steps

1. Run `wrk -h` from neutral cwd (no git required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```

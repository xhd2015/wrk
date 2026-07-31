# Scenario

**Feature**: `wrk -h` documents `--pr`, `--title`, and `--comment`

```
workspace/ -> wrk -h
  -> exit 0
  -> help mentions --pr, --title, --comment
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

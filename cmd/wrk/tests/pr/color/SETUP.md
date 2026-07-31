# Scenario

**Feature**: `--pr` success tokens use green ANSI when `--color` forces color

```
# non-TTY harness + --color
linked wt + fake gh (no open PR)
  -> wrk --pr --title T --comment C --color
  -> stdout success tokens green; URL plain
```

## Steps

- Leaves reuse linked-feature fixtures + fake gh and append `--color`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```

# Scenario

**Feature**: incomplete create flags still error; empty title/comment when both flags are present

```
# incomplete create (not show; not comment-only P2)
wrk --pr --title T            # no --comment
  -> non-zero; stderr names --comment

# empty after trim (create path still requires non-empty values when flags present)
wrk --pr --title "   " --comment C
wrk --pr --title T --comment "   "
  -> non-zero; stderr names the empty flag

# P2 owns comment-only: do NOT assert --pr --comment C fails for missing --title
```

## Steps

- Leaves set incomplete/empty argv; git fixture optional (flag-layer may fail first).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```

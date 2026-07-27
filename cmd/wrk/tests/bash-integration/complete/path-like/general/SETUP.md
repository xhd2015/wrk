# Scenario

**Feature**: general (positional) path-like cur yields empty custom completion

```
# not after --done: first positional is path-like
wrk --bash-integration --complete -- wrk <path-like> 1 -> empty stdout
# covers ./ ../ and absolute / prefixes
```

## Steps

1. Run three complete invocations: relative-dot, parent-relative, absolute.
2. Projects remain seeded from parent path-like setup.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CompleteCases = []CompleteCase{
		{Name: "relative-dot", Words: []string{"wrk", "./ex"}, CWord: 1},
		{Name: "parent-relative", Words: []string{"wrk", "../foo"}, CWord: 1},
		{Name: "absolute", Words: []string{"wrk", "/tmp/x"}, CWord: 1},
	}
	return nil
}
```

# Scenario

**Feature**: bare `--pr` hard-refuses when preconditions fail (linked wt, github origin, gh, attached HEAD)

```
# refuse gates
main repo cwd / non-github origin / detached HEAD / gh missing
  -> wrk --pr --title T --comment C
  -> non-zero; clear stderr; no PR create
```

## Steps

- Leaves seed the failing precondition and set default `--pr` argv.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```

# Scenario

**Feature**: wrk --main --cd rejects unexpected positionals

```
wrk --main --cd /path -> non-zero; unexpected arguments
```

## Steps

- Descendants set extra path arg with compose flags.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	return nil
}
```

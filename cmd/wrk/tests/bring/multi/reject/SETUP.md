# Scenario

**Feature**: multi-bring CLI rejections (exec, positionals, exact duplicates)

```
# multi + --exec          -> non-zero; --exec only with single --bring
# --bring p1 p2           -> non-zero; unexpected arguments (no multi-value sugar)
# --bring p1 --bring p1   -> non-zero; exact duplicate resolved path rejected
```

## Steps

- Leaves set `req.InProcess = true` and assert non-zero + stable stderr substrings.
- Success fixtures are only built when needed to prove no false GREEN (e.g. valid deps for duplicate/exec).

## Context

- Preferred exec error: `wrk: --exec is only valid with a single --bring path`
- Preferred positional: `wrk: unexpected arguments`
- Preferred duplicate: error naming duplicate / already listed / same path (soft wording).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```

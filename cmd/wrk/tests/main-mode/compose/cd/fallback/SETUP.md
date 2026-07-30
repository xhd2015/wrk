# Scenario

**Feature**: wrk --main --cd without follow-up uses runCd fallback (hint + path + shell)

```
WRK_FOLLOWUP_FILE unset
fake bash on PATH
wrk --main --cd
  -> stderr: wrk --bash-integration --install
  -> stdout: main\n
  -> LoginInteractive at main
```

## Preconditions

- Every fallback leaf **must** call `installFakeBash`.

## Steps

1. Create layout; install fake bash; do **not** enable follow-up channel.
2. Run compose `--main --cd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves install fake bash and set layout.
	return nil
}
```

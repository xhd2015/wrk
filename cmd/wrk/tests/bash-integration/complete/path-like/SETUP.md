# Scenario

**Feature**: path-like completion tokens yield empty custom candidates

```
# path-like cur: starts with / or ./ or ../
wrk --bash-integration --complete -- <words> <cword> -> empty stdout (exit 0)
# bash _wrk yields filename completion (script surface covered under print-script/install)
```

## Preconditions

- Path-like detection is prefix-based on the current word only (`/`, `./`, `../`).
- Seeded projects must not change path-like empty behavior (no inventing basenames).

## Steps

1. Seed standard projects so accidental basename leakage would be visible.
2. Descendants set path-like `CompleteWords` / `CompleteCWord`.

## Context

- Empty custom list is intentional: compspec `-o default` / `compopt -o default` then supply filenames.
- Non-path-like regression lives under complete/basenames, complete/flags, complete/dep.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	return nil
}

func assertPathLikeCompleteEmpty(t *testing.T, resp *Response) {
	t.Helper()
	assertCompleteExitOK(t, resp)
	if resp.Stdout != "" {
		t.Fatalf("path-like cur must yield empty --complete stdout; got %q", resp.Stdout)
	}
}
```

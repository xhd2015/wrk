# Scenario

**Feature**: wrapper create auto-cd controls (home-gated default; target-dir never)

```
# StartDir = FakeHome (user home) + auto-cd on + default create
source bash.sh; wrk <mainRepo> -> stderr "cd <wt>"; FinalPWD = worktree

# StartDir = main repo (not home)
source bash.sh; wrk -> no stderr cd; FinalPWD stays main

# StartDir = FakeHome + explicit <target-dir>
source bash.sh; wrk <mainRepo> <target>
  -> stdout path; no stderr cd; FinalPWD stays FakeHome

# StartDir = main + --force-cd: binary writes follow-up despite non-home
source bash.sh; wrk --force-cd -> stderr "cd <wt>"; FinalPWD = worktree

# WRK_AUTO_CD=0 / --no-cd suppress regardless of home
```

## Steps

1. Descendants seed main repo and choose StartDir (FakeHome vs mainRepo).
2. Target-dir leaf passes a second absolute positional path; shell stays put.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "wrapper")
	return nil
}
```

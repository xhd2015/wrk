# Scenario

**Feature**: Main checkout on detached HEAD is a fatal sync error

```
# main repo checkout --detach
myrepo (detached) -> wrk --sync -> non-zero; stderr mentions detached / named branch
```

## Steps

1. Init main-only repo with one commit.
2. `git checkout --detach` on main.
3. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainOnlyRepo(t, req)
	runGitIsolated(t, mainRepo, "checkout", "--detach")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.Args = []string{"--sync"}
	return nil
}
```

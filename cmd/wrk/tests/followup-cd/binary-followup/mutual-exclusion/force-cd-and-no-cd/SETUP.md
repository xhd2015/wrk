# Scenario

**Feature**: create with both --force-cd and --no-cd is rejected

```
myrepo (cwd) + WRK_FOLLOWUP_FILE
wrk --force-cd --no-cd -> non-zero; empty follow-up; error mentions both flags / mutual exclusion
```

## Steps

1. Init main repo (valid create context so rejection is about flags, not missing repo).
2. Run with both flags and follow-up env set.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--force-cd", "--no-cd"}
	return nil
}
```

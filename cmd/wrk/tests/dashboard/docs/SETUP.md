# Scenario

**Feature**: P4 docs coverage — help and skill document dashboard + `--new`

```
wrk -h
  -> mentions dashboard (bare no-args) and --new (create entry)

wrk skill --show
  -> embedded SKILL.md guides create via --new and bare → dashboard
```

## Steps

- Leaves run help / skill --show only (no git required).
- Coverage backfill where product already documents help; skill may RED until SKILL.md is updated.

```go
import "strings"

func Setup(t *testing.T, req *Request) error {
	// Docs leaves do not require a git checkout; neutral WorkRoot cwd.
	req.RepoDir = req.WorkRoot
	_ = assertEmbeddedSkillStdoutDash
	return nil
}

// assertEmbeddedSkillStdoutDash mirrors skill/ tree helper (not shared across feature dirs).
func assertEmbeddedSkillStdoutDash(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "WRK_SKILL_DOCTEST_MARKER") {
		t.Fatalf("skill stdout missing WRK_SKILL_DOCTEST_MARKER; got %q", stdout)
	}
	if !strings.Contains(stdout, "name: wrk") {
		t.Fatalf("skill stdout missing YAML name: wrk; got %q", stdout)
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("skill stdout must end with trailing newline")
	}
}
```

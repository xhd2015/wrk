# Scenario

**Feature**: `--exec` after successful `--cd` runs the command in the jump directory

```
# in-place --cd (WRK_FOLLOWUP_FILE set) + --exec
WRK_FOLLOWUP_FILE open
wrk --cd <abs> --exec pwd
  -> write follow-up: cd <abs>\n
  -> run pwd with cmd.Dir=<abs>; child stdout on process stdout
  -> exit 0; still no interactive shell
```

## Preconditions

- Leaves open the follow-up channel (`UseFollowupEnv` + `FollowupFile`).
- Target dir need not be a git repo.

## Steps

- Create jump target under WorkRoot; set `wrk --cd <abs> --exec …`.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureExecCDHelpersUsed()
	return nil
}

func enableExecFollowup(t *testing.T, req *Request) {
	t.Helper()
	req.FollowupFile = filepath.Join(req.WorkRoot, "followup.txt")
	req.UseFollowupEnv = true
}

func execCDTarget(t *testing.T, req *Request, name string) string {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, name)
	mkdirAll(t, dir)
	return resolveAbs(t, dir)
}

func assertFollowupCDExact(t *testing.T, req *Request, wantAbs string) {
	t.Helper()
	wantAbs = resolveAbs(t, wantAbs)
	data, err := os.ReadFile(req.FollowupFile)
	if err != nil {
		t.Fatalf("read follow-up %s: %v", req.FollowupFile, err)
	}
	want := "cd " + wantAbs + "\n"
	if string(data) != want {
		t.Fatalf("follow-up:\n want %q\n  got %q", want, string(data))
	}
}

func ensureExecCDHelpersUsed() {
	_ = enableExecFollowup
	_ = execCDTarget
	_ = assertFollowupCDExact
}
```

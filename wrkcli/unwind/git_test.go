package unwind

import (
	"bytes"
	"io"
	"testing"
)

func TestGitRunDirIOUsesHostIO(t *testing.T) {
	var gotOut, gotErr io.Writer
	var gotArgs []string
	restore := useHost(Host{
		GitRunIO: func(repo string, stdout, stderr io.Writer, args ...string) error {
			gotOut, gotErr = stdout, stderr
			gotArgs = append([]string{repo}, args...)
			return nil
		},
	})
	defer restore()

	var buf bytes.Buffer
	hio := HostIO{Stdout: &buf, Stderr: &buf}
	if err := gitRunDirIO("/tmp/repo", hio, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if gotOut != &buf || gotErr != &buf {
		t.Fatalf("stdio not redirected: out=%v err=%v", gotOut, gotErr)
	}
	if len(gotArgs) != 3 || gotArgs[0] != "/tmp/repo" || gotArgs[1] != "add" || gotArgs[2] != "-A" {
		t.Fatalf("args=%v", gotArgs)
	}
}

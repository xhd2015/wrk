## Expected

- Exit code 0.
- Root scope skips (`no changes`); `sub/` plans `sub/v0.2.4`.
- Tag `sub/v0.2.4` created at HEAD; root release tag count unchanged.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	short := shortHEAD(t, req.MainRepo)
	wantStdout := fmt.Sprintf(
		"v0.0.1        no changes                     ->  skip\nsub/v0.2.3    owned changed                  ->  sub/v0.2.4\ntagged sub/v0.2.4 @ %s\n1 tag created\n",
		short,
	)
	assert.Output(t, resp.Stdout, tagNextStdoutV2(wantStdout))

	if !tagRefExists(t, req.MainRepo, "sub/v0.2.4") {
		t.Fatal("sub/v0.2.4 tag should exist after apply")
	}
	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("root v0.0.2 should not be created when only sub/ changed")
	}
}
```
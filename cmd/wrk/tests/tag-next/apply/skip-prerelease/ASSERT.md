## Expected Output

```
v0.0.2        prerelease head (v0.0.3-alpha)  ->  skip
0 tag created
```

## Expected

- Exit code 0.
- Stdout shows prerelease-head skip and `0 tag created`.
- No new release tag beyond existing tags.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantStdout := "v0.0.2        prerelease head (v0.0.3-alpha)  ->  skip\n0 tag created\n"
	assert.Output(t, resp.Stdout, tagNextStdoutV2(wantStdout))

	if tagRefExists(t, req.MainRepo, "v0.0.3") {
		t.Fatal("v0.0.3 release tag should not be created when prerelease head blocks bump")
	}
}
```
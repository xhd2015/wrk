package wrkcli

import (
	"bytes"
	"strings"
	"testing"
)

func TestIndentingWriterPrefixesLines(t *testing.T) {
	var buf bytes.Buffer
	w := newIndentingWriter(&buf, "      ")
	_, _ = w.Write([]byte("merged branch x into main\n"))
	_, _ = w.Write([]byte("line2\npartial"))
	_, _ = w.Write([]byte("-cont\n"))
	got := buf.String()
	want := "      merged branch x into main\n      line2\n      partial-cont\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestIndentingWriterEmptyWrite(t *testing.T) {
	var buf bytes.Buffer
	w := newIndentingWriter(&buf, "  ")
	n, err := w.Write(nil)
	if err != nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if buf.Len() != 0 {
		t.Fatalf("unexpected %q", buf.String())
	}
}

func TestComposeStageOutIndents(t *testing.T) {
	var out, errB bytes.Buffer
	st := &composeStageWriter{
		total:  2,
		indent: "      ",
		errW:   &errB,
		outW:   &out,
	}
	st.bodyOut = newIndentingWriter(st.outW, st.indent)
	st.bodyErr = newIndentingWriter(st.errW, st.indent)
	st.mark(1, "merge-back")
	_, _ = st.Out().Write([]byte("merged branch x into main\n"))
	if !strings.HasPrefix(errB.String(), "[1/2] merge-back\n") {
		t.Fatalf("marker: %q", errB.String())
	}
	if out.String() != "      merged branch x into main\n" {
		t.Fatalf("body: %q", out.String())
	}
}

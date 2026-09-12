package wrkcli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestComposeStageWriterMark(t *testing.T) {
	var errBuf, outBuf bytes.Buffer
	st := &composeStageWriter{
		total:  3,
		indent: "      ",
		errW:   &errBuf,
		outW:   &outBuf,
	}
	st.mark(1, "gen-commit-msg")
	st.mark(2, "merge-back")
	st.mark(3, "ship · concurrent")
	got := errBuf.String()
	if !strings.Contains(got, "[1/3] gen-commit-msg\n") {
		t.Fatalf("marker1: %q", got)
	}
	if !strings.Contains(got, "[2/3] merge-back\n") {
		t.Fatalf("marker2: %q", got)
	}
	if !strings.Contains(got, "[3/3] ship · concurrent\n") {
		t.Fatalf("marker3: %q", got)
	}
}

func TestNewLandComposeStagesIndexes(t *testing.T) {
	st, shipIdx := newLandComposeStages(true, true, false, false, true)
	if st.total != 3 {
		t.Fatalf("total=%d want 3", st.total)
	}
	if shipIdx != 3 {
		t.Fatalf("shipIdx=%d want 3", shipIdx)
	}
	st2, ship2 := newLandComposeStages(false, true, false, false, true)
	if st2.total != 2 || ship2 != 2 {
		t.Fatalf("no-commit: total=%d ship=%d", st2.total, ship2)
	}
}

func TestIsContextCanceled(t *testing.T) {
	if !isContextCanceled(context.Canceled) {
		t.Fatal("expected context.Canceled")
	}
	if isContextCanceled(errors.New("boom")) {
		t.Fatal("plain error must not match")
	}
}

func TestShipFailFastCancelsSibling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var started, sawCancel atomic.Int32
	done := make(chan struct{})
	go func() {
		started.Add(1)
		<-ctx.Done()
		sawCancel.Add(1)
		close(done)
	}()

	for started.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sibling was not cancelled")
	}
	if sawCancel.Load() != 1 {
		t.Fatalf("sawCancel=%d", sawCancel.Load())
	}
}

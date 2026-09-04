package web

import (
	"net"
	"strings"
	"testing"
)

func TestListenLocalSkipsBusyWildcard8080(t *testing.T) {
	holder, err := net.Listen("tcp", ":8080")
	if err == nil {
		t.Cleanup(func() { _ = holder.Close() })
	}
	// If err != nil, something else already holds *:8080 — auto-pick must still skip it.

	ln, port, err := listenLocal(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	if port == 8080 {
		t.Fatal("listenLocal(0) claimed 8080 while :8080 is busy")
	}
}

func TestListenLocalExplicitPortBusy(t *testing.T) {
	holder, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = holder.Close() })
	p := holder.Addr().(*net.TCPAddr).Port

	_, _, err = listenLocal(p)
	if err == nil {
		t.Fatalf("listenLocal(%d) succeeded while :%d is held", p, p)
	}
	if !strings.Contains(err.Error(), "listen") {
		t.Fatalf("error=%v", err)
	}
}

func TestListenLocalPrefers8080WhenFree(t *testing.T) {
	probe, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Skipf("8080 busy: %v", err)
	}
	_ = probe.Close()

	ln, port, err := listenLocal(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	if port != 8080 {
		t.Fatalf("listenLocal(0)=%d want 8080 when free", port)
	}
}

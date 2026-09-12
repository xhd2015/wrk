package wrkcli

import (
	"io"
	"os"
	"sync"
)

// runWithWritersAsStdio temporarily points os.Stdout/os.Stderr at pipes that
// copy into out/errW, runs fn, then restores. Use only for libraries that
// hardcode os.Stdout/Stderr until they accept writers.
//
// Not safe if fn starts goroutines that keep writing after return.
func runWithWritersAsStdio(out, errW io.Writer, fn func() error) error {
	if out == nil {
		out = os.Stdout
	}
	if errW == nil {
		errW = os.Stderr
	}

	rOut, wOut, err := os.Pipe()
	if err != nil {
		return err
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		_ = rOut.Close()
		_ = wOut.Close()
		return err
	}

	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(out, rOut)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(errW, rErr)
	}()

	fnErr := fn()

	_ = wOut.Close()
	_ = wErr.Close()
	wg.Wait()
	_ = rOut.Close()
	_ = rErr.Close()

	os.Stdout, os.Stderr = oldOut, oldErr
	return fnErr
}

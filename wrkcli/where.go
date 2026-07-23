package wrkcli

import (
	"fmt"
	"io"
	"os"

	"github.com/xhd2015/wrk/wrkcli/storage"
)

func runWhere(out io.Writer, wrkHome, basename string) error {
	if out == nil {
		out = os.Stdout
	}
	if !isBasename(basename) {
		return fmt.Errorf("wrk: --where requires a basename-only argument")
	}
	matches, err := storage.FindProjectsByBasename(wrkHome, basename)
	if err != nil {
		return err
	}
	switch len(matches) {
	case 0:
		return fmt.Errorf("wrk: no saved project for basename %q", basename)
	case 1:
		fmt.Fprintln(out, matches[0])
	default:
		for _, p := range matches {
			fmt.Fprintln(out, p)
		}
	}
	return nil
}
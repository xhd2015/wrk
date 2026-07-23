package wrkcli

import (
	"fmt"

	"github.com/xhd2015/wrk/wrkcli/storage"
)

func runWhere(wrkHome, basename string) error {
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
		fmt.Fprintln(cliStdout(), matches[0])
	default:
		for _, p := range matches {
			fmt.Fprintln(cliStdout(), p)
		}
	}
	return nil
}
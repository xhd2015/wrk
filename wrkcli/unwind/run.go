package unwind

// Run is the CLI entry for wrk --unwind. host supplies push/sync/gen-commit
// /reinstall (and verbose git) from wrkcli without an import cycle.
func Run(workDir, wrkHome string, flags UnwindFlags, host Host) error {
	restore := useHost(host)
	defer restore()
	return runUnwind(workDir, wrkHome, flags)
}

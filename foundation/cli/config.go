package cli

import (
	"github.com/spf13/pflag"
	"os"
)

var (
	ErrorExitCode        = 1 // Exit code when Command errors reported.
	UsageErrorExitCode   = 2 // Exit code when a UsageError is reported.
	LongHelpFlag         = "help"
	ShortHelpFlag        = "h"
	FlagsErrorHandling   = pflag.ContinueOnError
	FlagsInterspersed    bool
	DefaultPrinterOutput = os.Stderr
	testNoExit           bool
)

func testSettings() {
	testNoExit = true
	DefaultPrinterOutput = os.Stdout
}

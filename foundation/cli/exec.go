package cli

import (
	"errors"
)

var ErrNoExecution = errors.New("no execution attached to this command")

func executeErr(args []string, does CommandFunc, flags *Flags, out *Printer) error {
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.usageRequested() {
		flags.Usage()
		return nil
	}
	if does == nil {
		return ErrNoExecution
	}
	if err := runGlobalPreExec(); err != nil {
		return checkUsageError(flags, out, err)
	}
	if err := does(flags, out); err != nil {
		return checkUsageError(flags, out, err)
	}
	return nil
}

func executeExit(args []string, does CommandFunc, flags *Flags, out *Printer) {
	if err := executeErr(args, does, flags, out); err != nil {
		out.Fatalln(err.Error())
	}
}

func checkUsageError(flags *Flags, out *Printer, err error) error {
	var uerr *UsageError
	if errors.As(err, &uerr) {
		out.fatalUsage(flags.Usage, err)
	}
	return err
}

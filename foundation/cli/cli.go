package cli

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"path/filepath"
	"strings"
)

type Flags struct {
	*flag.FlagSet
}

func (f *Flags) setHelpFlag() {
	if len(ShortHelpFlag) == 0 && len(LongHelpFlag) == 0 {
		LongHelpFlag = "help"
		ShortHelpFlag = "h"
	}
	desc := "Prints this usage information"
	switch {
	case len(LongHelpFlag) == 0:
		f.BoolP("help", ShortHelpFlag, false, desc)
	case len(ShortHelpFlag) == 0:
		f.Bool(LongHelpFlag, false, desc)
	default:
		f.BoolP(LongHelpFlag, ShortHelpFlag, false, desc)
	}
}

func (f *Flags) usageRequested() bool {
	if val, err := f.GetBool(LongHelpFlag); err == nil && val {
		return true
	}
	if val, err := f.GetBool(ShortHelpFlag); err == nil && val {
		return true
	}
	return false
}

// CommandFunc is a function that may be executed within a [Command].
type CommandFunc = func(flags *Flags, out *Printer) error

// Command is an executable function in a CLI.
// It should be linked to a [Command] to establish a tree of commands available to the user.
type Command struct {
	subcommands
	flags         *Flags
	exec          CommandFunc
	name          string
	parent        string
	summary       string
	usageExamples []string
	prose         string
	printer       *Printer
	aliases       []string
	isTLC         bool
}

func TopLevelCommand() *Command {
	base := filepath.Base(os.Args[0])
	out := NewPrinter()
	return topLevelCommand(base, out)
}

func TopLevelCommandWithName(name string) *Command {
	name = cleanseName(name)
	out := NewPrinter()
	return topLevelCommand(name, out)
}

func topLevelCommand(name string, out *Printer) *Command {
	cmd, err := newCommand(cleanseName(name), "", "", out)
	if err != nil {
		out.Fatalln("Failed to create top level command:", err)
	}
	return cmd
}

func newCommand(name, parent, summary string, printer *Printer) (*Command, error) {
	_name := cleanseName(name)
	if len(name) == 0 {
		return nil, fmt.Errorf("normalized name '%s' is blank", name)
	}
	summary = strings.TrimSpace(summary)
	fs := flag.NewFlagSet(_name, FlagsErrorHandling)
	flags := &Flags{FlagSet: fs}
	flags.setHelpFlag()
	flags.SetInterspersed(FlagsInterspersed)
	cmd := &Command{
		flags:   flags,
		name:    _name,
		parent:  parent,
		isTLC:   len(parent) == 0,
		summary: summary,
		printer: printer,
	}
	flags.Usage = cmd.setupUsage(flags)
	return cmd, nil
}

// Does specifies the [CommandFunc] that should be executed by this [Command].
func (c *Command) Does(commandFunc CommandFunc) *Command {
	if commandFunc == nil {
		return c
	}
	c.exec = commandFunc
	return c
}

// CommandPath returns the reference chain for this [Command].
func (c *Command) CommandPath() string {
	if len(c.parent) == 0 {
		return c.name
	}
	return fmt.Sprintf("%s %s", c.parent, c.name)
}

// Flags returns the [Flags] for this [Command].
func (c *Command) Flags() *Flags {
	return c.flags
}

// AddUsageExample adds an example of how the Command can be invoked, using placeholders for arguments and flags.
//
// Example:
//
//	FILE [MODE] [FLAGS]
//
// Parent commands and this command name will be used to prefix examples.
func (c *Command) AddUsageExample(example string) {
	c.usageExamples = append(c.usageExamples, example)
}

// SetUsageProse allows specifying a longer description of the [Command] that will be output when a usage information is requested.
func (c *Command) SetUsageProse(format string, args ...any) *Command {
	c.prose = fmt.Sprintf(format, args...)
	return c
}

// Exec executes the command with given arguments, parsing flags.
func (c *Command) Exec(args []string) error {
	if len(args) > 0 {
		cmd, ok := c.resolveCommand(args[0])
		if ok {
			return cmd.Exec(args[1:])
		}
	}
	out := c.Printer()
	if c.isTLC && !testNoExit {
		executeExit(args, c.exec, c.flags, out)
		return nil
	}
	return executeErr(args, c.exec, c.flags, out)
}

func (c *Command) setupUsage(flags *Flags) func() {
	return func() {
		var (
			buf strings.Builder
			out = c.printer
		)
		addSep := func() func() {
			prevPrint := false
			return func() {
				if prevPrint {
					buf.WriteString("\n\n")
				}
				prevPrint = true
			}
		}()
		if len(c.summary) > 0 {
			addSep()
			buf.WriteString(c.summary)
		}
		if len(c.usageExamples) > 0 {
			addSep()
			for i, example := range c.usageExamples {
				if i == 0 {
					buf.WriteString(fmt.Sprintf("USAGE: %s %s", c.CommandPath(), example))
				} else {
					buf.WriteString(fmt.Sprintf("\n       %s %s", c.CommandPath(), example))
				}
			}
		}
		if len(c.prose) > 0 {
			addSep()
			buf.WriteString(c.prose)
		}
		if len(c.aliases) > 0 {
			addSep()
			buf.WriteString(fmt.Sprintf("ALIASES:\n\t%s\n", strings.Join(c.aliases, ", ")))
		}
		if flags.HasFlags() {
			addSep()
			buf.WriteString("FLAGS:\n")
			buf.WriteString(flags.FlagUsages())
		}
		if len(c.cmds) > 0 {
			addSep()
			buf.WriteString("SUBCOMMANDS:\n")
			for name, cmd := range c.cmds {
				names := strings.Join(append([]string{name}, cmd.aliases...), ", ")
				buf.WriteString(fmt.Sprintf("\t%s   %s\n", names, cmd.summary))
			}
		}
		out.Print(buf.String())
	}
}

// AddCommand adds a sub-command to this [Command].
// The name parameter will be cleansed to remove spaces, and normalize to lower-case.
// Aliases may be added as a way to support shorter variants of the same [Command].
func (c *Command) AddCommand(name, summary string, aliases ...string) *Command {
	_name := cleanseName(name)
	out := c.Printer()
	if len(name) == 0 {
		out.Fatalf("Invalid command name '%s'\n", name)
		panic("invalid command name")
	}
	_summary := strings.TrimSpace(summary)
	if len(_summary) == 0 {
		out.Fatalf("Command summary for '%s' is required\n", name)
		panic("missing summary")
	}
	cmd, err := newCommand(_name, c.CommandPath(), _summary, out)
	if err != nil {
		out.Fatalf("Failed to create command '%s': %v\n", _name, err)
		panic("failed to create command")
	}
	if err := c.addSubCommand(name, cmd, aliases...); err != nil {
		out.Fatalf("Failed to add command '%s': %v", _name, err)
		panic("failed to add sub-command")
	}
	return cmd
}

// Printer returns the cached [Printer] for this [Command].
func (c *Command) Printer() *Printer {
	if c.printer == nil {
		c.printer = NewPrinter()
	}
	return c.printer
}

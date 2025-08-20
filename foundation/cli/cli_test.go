package cli

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	testSettings()
}

func TestCommand_Exec(t *testing.T) {
	set := TopLevelCommandWithName("testing")
	assert.ErrorIs(t, set.Exec(nil), ErrNoExecution)

	cmd := set.AddCommand("test", "test command")
	assert.ErrorIs(t, set.Exec([]string{"test"}), ErrNoExecution)

	executed := false
	cmd.Does(func(flags *Flags, _ *Printer) error {
		executed = true
		return nil
	})
	assert.NoError(t, set.Exec([]string{"test"}))
	assert.True(t, executed)

	assert.ErrorIs(t, set.Exec([]string{"Does", "not", "exist"}), ErrNoExecution)
}

func TestCommand_AddSubCommand(t *testing.T) {
	cmdExecuted := 0
	subExecuted := 0
	cmd := testCommandWithSubcommand(t, &cmdExecuted, &subExecuted)

	assert.NoError(t, cmd.Exec([]string{"-h"}))

	cmd = testCommandWithSubcommand(t, &cmdExecuted, &subExecuted)
	assert.NoError(t, cmd.Exec([]string{"blah"}), "Should execute test without error")
	assert.Equal(t, 1, cmdExecuted)
	assert.Equal(t, 0, subExecuted)

	cmd = testCommandWithSubcommand(t, &cmdExecuted, &subExecuted)
	assert.NoError(t, cmd.Exec([]string{"SUB"}))
	assert.Equal(t, 1, cmdExecuted)
	assert.Equal(t, 1, subExecuted)
}

func TestCommand_AddCommand_Aliases(t *testing.T) {
	cmdExecuted := 0
	subExecuted := 0
	cmd := testCommand(t, &cmdExecuted, &subExecuted)
	assert.NoError(t, cmd.Exec([]string{"test", "a"}), "Should execute test without error")
	assert.Equal(t, 0, cmdExecuted)
	assert.Equal(t, 1, subExecuted)

	assert.NoError(t, cmd.Exec([]string{"test", "b"}), "Should execute test without error")
	assert.Equal(t, 0, cmdExecuted)
	assert.Equal(t, 2, subExecuted)
}

func TestPrinter_RespondUsage(t *testing.T) {
	cmdExecuted := 0
	subExecuted := 0
	cmd := testCommand(t, &cmdExecuted, &subExecuted)
	tmp := os.Args
	t.Cleanup(func() {
		os.Args = tmp
	})
	os.Args = []string{"test", ShortHelpFlag, "something", "else"}
	err := cmd.Exec(os.Args)
	require.NoError(t, err)
}

func testCommand(t *testing.T, cmdExecuted, subExecuted *int) *Command {
	t.Helper()
	set := TopLevelCommandWithName("testing")
	cmd := set.AddCommand("test", "test command", "t")
	cmd.Flags().String("message", "", "Sets a message")
	cmd.Does(func(_ *Flags, _ *Printer) error {
		*cmdExecuted++
		return nil
	})

	sub := cmd.AddCommand("sub", "test subcommand", "a", "b")
	require.Equal(t, "testing test", sub.parent)
	sub.Does(func(flags *Flags, _ *Printer) error {
		*subExecuted++
		return nil
	})
	return set
}

func testCommandWithSubcommand(t *testing.T, cmdExecuted, subExecuted *int) *Command {
	cmd, err := newCommand("test", "", "test command", NewPrinter())
	require.NoError(t, err)
	cmd.Does(func(flags *Flags, _ *Printer) error {
		*cmdExecuted++
		return nil
	})
	cmd.Flags().String("message", "", "Sets a message")

	sub := cmd.AddCommand("sub", "test subcommand", "a", "b")
	assert.Equal(t, "test", sub.parent)
	sub.Does(func(flags *Flags, _ *Printer) error {
		*subExecuted++
		return nil
	})
	return cmd
}

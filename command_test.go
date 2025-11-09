package command_test

import (
	"testing"

	"github.com/gloo-foo/testable/assertion"
	"github.com/gloo-foo/testable/run"
	command "github.com/yupsh/ls"
)

func TestLs_Current(t *testing.T) {
	result := run.Quick(command.Ls("."))
	assertion.NoError(t, result.Err)
	// Should list current directory
}

func TestLs_All(t *testing.T) {
	result := run.Quick(command.Ls(".", command.AllFiles))
	assertion.NoError(t, result.Err)
}

func TestLs_Long(t *testing.T) {
	result := run.Quick(command.Ls(".", command.LongFormat))
	assertion.NoError(t, result.Err)
}

func TestLs_Human(t *testing.T) {
	result := run.Quick(command.Ls(".", command.HumanReadable))
	assertion.NoError(t, result.Err)
}

func TestLs_Recursive(t *testing.T) {
	result := run.Quick(command.Ls(".", command.Recursive))
	assertion.NoError(t, result.Err)
}


package ls_test

import (
	"context"
	"os"

	"github.com/yupsh/ls"
	"github.com/yupsh/ls/opt"
)

func ExampleLs() {
	ctx := context.Background()

	cmd := ls.Ls(".")
	cmd.Execute(ctx, nil, os.Stdout, os.Stderr)
	// Output will vary based on directory contents
}

func ExampleLs_longFormat() {
	ctx := context.Background()

	cmd := ls.Ls(".", opt.LongFormat, opt.HumanReadable)
	cmd.Execute(ctx, nil, os.Stdout, os.Stderr)
	// Output will vary based on directory contents
}

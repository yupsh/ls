package ls

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	yup "github.com/yupsh/framework"
	"github.com/yupsh/framework/opt"

	localopt "github.com/yupsh/ls/opt"
)

// Flags represents the configuration options for the ls command
type Flags = localopt.Flags

// Command implementation
type command opt.Inputs[string, Flags]

// Ls creates a new ls command with the given parameters
func Ls(parameters ...any) yup.Command {
	return command(opt.Args[string, Flags](parameters...))
}

func (c command) Execute(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	// Check for cancellation before starting
	if err := yup.CheckContextCancellation(ctx); err != nil {
		return err
	}

	paths := c.Positional
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for i, path := range paths {
		// Check for cancellation before each path
		if err := yup.CheckContextCancellation(ctx); err != nil {
			return err
		}

		if len(paths) > 1 && i > 0 {
			fmt.Fprintln(stdout)
		}

		if len(paths) > 1 {
			fmt.Fprintf(stdout, "%s:\n", path)
		}

		if err := c.listPath(ctx, path, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "ls: %s: %v\n", path, err)
		}
	}

	return nil
}

func (c command) listPath(ctx context.Context, path string, output, stderr io.Writer) error {
	// Check for cancellation before starting
	if err := yup.CheckContextCancellation(ctx); err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		// Single file
		c.printFile(output, path, info)
		return nil
	}

	// Directory listing
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Check for cancellation after reading directory
	if err := yup.CheckContextCancellation(ctx); err != nil {
		return err
	}

	// Filter and sort entries
	var items []os.DirEntry
	for i, entry := range entries {
		// Check for cancellation periodically (every 1000 entries for efficiency)
		if i%1000 == 0 {
			if err := yup.CheckContextCancellation(ctx); err != nil {
				return err
			}
		}

		// Skip hidden files unless -a flag is set
		if !bool(c.Flags.AllFiles) && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		items = append(items, entry)
	}

	// Sort entries
	c.sortEntries(ctx, items, path)

	// Check for cancellation after sorting
	if err := yup.CheckContextCancellation(ctx); err != nil {
		return err
	}

	// Print entries
	for i, entry := range items {
		// Check for cancellation periodically (every 100 entries for efficiency)
		if i%100 == 0 {
			if err := yup.CheckContextCancellation(ctx); err != nil {
				return err
			}
		}

		entryPath := filepath.Join(path, entry.Name())
		entryInfo, err := entry.Info()
		if err != nil {
			fmt.Fprintf(stderr, "ls: %s: %v\n", entryPath, err)
			continue
		}

		c.printFile(output, entryPath, entryInfo)

		// Recursive listing if requested
		if bool(c.Flags.Recursive) && entryInfo.IsDir() {
			fmt.Fprintln(output)
			fmt.Fprintf(output, "%s:\n", entryPath)
			c.listPath(ctx, entryPath, output, stderr)
		}
	}

	return nil
}

func (c command) sortEntries(ctx context.Context, entries []os.DirEntry, basePath string) {
	// Note: sort.Slice doesn't support context cancellation directly,
	// but we can check for cancellation before starting the sort operation.
	// For very large directories, this provides a cancellation point.
	if err := yup.CheckContextCancellation(ctx); err != nil {
		return // Return early on cancellation, leaving entries unsorted
	}

	sort.Slice(entries, func(i, j int) bool {
		less := false

		switch c.Flags.SortBy {
		case localopt.SortByTime:
			info1, _ := entries[i].Info()
			info2, _ := entries[j].Info()
			less = info1.ModTime().Before(info2.ModTime())
		case localopt.SortBySize:
			info1, _ := entries[i].Info()
			info2, _ := entries[j].Info()
			less = info1.Size() < info2.Size()
		default: // SortByName
			less = entries[i].Name() < entries[j].Name()
		}

		if bool(c.Flags.Reverse) {
			less = !less
		}

		return less
	})
}

func (c command) printFile(output io.Writer, path string, info os.FileInfo) {
	if bool(c.Flags.LongFormat) {
		c.printLongFormat(output, path, info)
	} else {
		fmt.Fprintln(output, filepath.Base(path))
	}
}

func (c command) printLongFormat(output io.Writer, path string, info os.FileInfo) {
	mode := info.Mode()
	size := info.Size()
	modTime := info.ModTime().Format(time.Stamp)
	name := filepath.Base(path)

	// Format size
	sizeStr := fmt.Sprintf("%d", size)
	if bool(c.Flags.HumanReadable) {
		sizeStr = c.humanReadableSize(size)
	}

	fmt.Fprintf(output, "%s %8s %s %s\n", mode, sizeStr, modTime, name)
}

func (c command) humanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}

	units := []string{"K", "M", "G", "T"}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f%s", float64(size)/float64(div), units[exp])
}

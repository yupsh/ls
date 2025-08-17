package command

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	yup "github.com/gloo-foo/framework"
)

type command yup.Inputs[string, flags]

func Ls(parameters ...any) yup.Command {
	return command(yup.Initialize[string, flags](parameters...))
}

func (p command) Executor() yup.CommandExecutor {
	return func(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
		patterns := p.Positional
		if len(patterns) == 0 {
			patterns = []string{"."}
		}

		// Expand globs and braces
		paths, err := p.expandPatterns(patterns)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "ls: %v\n", err)
			return err
		}

		// If no paths matched, use the original patterns
		if len(paths) == 0 {
			paths = patterns
		}

		// Process each path
		for _, path := range paths {
			// Get file info
			info, err := os.Stat(path)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "ls: %s: %v\n", path, err)
				continue
			}

			// If it's a file, just format and output it
			if !info.IsDir() {
				output := p.formatEntry(path, info)
				_, _ = fmt.Fprintln(stdout, output)
				continue
			}

			// It's a directory - list contents
			if err := p.listDirectory(path, stdout, stderr); err != nil {
				_, _ = fmt.Fprintf(stderr, "ls: %s: %v\n", path, err)
			}
		}

		return nil
	}
}

// expandBraces expands brace expressions like {a,b,c} or {1..10}
func expandBraces(pattern string) []string {
	// Find the first opening brace
	start := strings.Index(pattern, "{")
	if start == -1 {
		// No braces, return as-is
		return []string{pattern}
	}

	// Find the matching closing brace
	depth := 0
	end := -1
	for i := start; i < len(pattern); i++ {
		if pattern[i] == '{' {
			depth++
		} else if pattern[i] == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		// No matching closing brace, return as-is
		return []string{pattern}
	}

	// Extract prefix, brace content, and suffix
	prefix := pattern[:start]
	braceContent := pattern[start+1 : end]
	suffix := pattern[end+1:]

	var expansions []string

	// Check if it's a range expression like {1..10} or {a..z}
	if strings.Contains(braceContent, "..") {
		parts := strings.Split(braceContent, "..")
		if len(parts) == 2 {
			expansions = expandRange(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// If not a range, treat as comma-separated list
	if len(expansions) == 0 {
		items := strings.Split(braceContent, ",")
		for _, item := range items {
			expansions = append(expansions, strings.TrimSpace(item))
		}
	}

	// Build result by combining prefix, each expansion, and suffix
	var result []string
	for _, exp := range expansions {
		combined := prefix + exp + suffix
		// Recursively expand any remaining braces in the result
		result = append(result, expandBraces(combined)...)
	}

	return result
}

// expandRange expands a range like 1..10 or a..z
func expandRange(start, end string) []string {
	var result []string

	// Try numeric range
	var s, e int
	if _, err1 := fmt.Sscanf(start, "%d", &s); err1 == nil {
		if _, err2 := fmt.Sscanf(end, "%d", &e); err2 == nil {
			if s <= e {
				for i := s; i <= e; i++ {
					result = append(result, fmt.Sprintf("%d", i))
				}
			} else {
				for i := s; i >= e; i-- {
					result = append(result, fmt.Sprintf("%d", i))
				}
			}
			return result
		}
	}

	// Try character range
	if len(start) == 1 && len(end) == 1 {
		s, e := rune(start[0]), rune(end[0])
		if s <= e {
			for r := s; r <= e; r++ {
				result = append(result, string(r))
			}
		} else {
			for r := s; r >= e; r-- {
				result = append(result, string(r))
			}
		}
		return result
	}

	// If neither works, return the original as comma-separated
	return []string{start, end}
}

// expandPatterns expands glob patterns and brace expressions
func (p command) expandPatterns(patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// First expand braces
		expanded := expandBraces(pattern)

		// Then expand globs for each expanded pattern
		for _, exp := range expanded {
			matches, err := filepath.Glob(exp)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %s: %v", exp, err)
			}

			// If no matches, keep the original pattern
			if len(matches) == 0 {
				if !seen[exp] {
					result = append(result, exp)
					seen[exp] = true
				}
			} else {
				for _, match := range matches {
					if !seen[match] {
						result = append(result, match)
						seen[match] = true
					}
				}
			}
		}
	}

	return result, nil
}

// listDirectory lists the contents of a directory
func (p command) listDirectory(dirPath string, stdout, stderr io.Writer) error {
	if p.Flags.Recursive {
		return p.listRecursive(dirPath, stdout, stderr)
	}
	return p.listSingle(dirPath, stdout, stderr)
}

// listSingle lists a single directory
func (p command) listSingle(dirPath string, stdout, stderr io.Writer) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Filter entries
	var filtered []fs.DirEntry
	for _, entry := range entries {
		// Skip hidden files unless AllFiles is set
		if !bool(p.Flags.AllFiles) && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		filtered = append(filtered, entry)
	}

	// Sort entries
	p.sortEntries(filtered)

	// Format and output entries
	for _, entry := range filtered {
		fullPath := filepath.Join(dirPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "ls: %s: %v\n", fullPath, err)
			continue
		}
		line := p.formatEntry(fullPath, info)
		_, _ = fmt.Fprintln(stdout, line)
	}

	return nil
}

// listRecursive lists a directory recursively
func (p command) listRecursive(dirPath string, stdout, stderr io.Writer) error {
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "ls: %s: %v\n", path, err)
			return nil
		}

		// Skip hidden files/dirs unless AllFiles is set
		if !bool(p.Flags.AllFiles) && strings.HasPrefix(d.Name(), ".") && path != dirPath {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "ls: %s: %v\n", path, err)
			return nil
		}

		line := p.formatEntry(path, info)
		_, _ = fmt.Fprintln(stdout, line)

		return nil
	})

	return err
}

// sortEntries sorts directory entries based on flags
func (p command) sortEntries(entries []fs.DirEntry) {
	sortBy := string(p.Flags.SortBy)

	sort.Slice(entries, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "time":
			infoI, errI := entries[i].Info()
			infoJ, errJ := entries[j].Info()
			if errI != nil || errJ != nil {
				less = entries[i].Name() < entries[j].Name()
			} else {
				less = infoI.ModTime().After(infoJ.ModTime())
			}
		case "size":
			infoI, errI := entries[i].Info()
			infoJ, errJ := entries[j].Info()
			if errI != nil || errJ != nil {
				less = entries[i].Name() < entries[j].Name()
			} else {
				less = infoI.Size() > infoJ.Size()
			}
		default: // "name" or empty
			less = entries[i].Name() < entries[j].Name()
		}

		if p.Flags.Reverse {
			return !less
		}
		return less
	})
}

// formatEntry formats a single file/directory entry
func (p command) formatEntry(path string, info fs.FileInfo) string {
	if p.Flags.LongFormat {
		return p.formatLong(path, info)
	}
	return filepath.Base(path)
}

// formatLong formats entry in long format
func (p command) formatLong(path string, info fs.FileInfo) string {
	mode := info.Mode()
	size := info.Size()
	modTime := info.ModTime().Format(time.DateTime)
	name := filepath.Base(path)

	// Format size
	sizeStr := fmt.Sprintf("%d", size)
	if p.Flags.HumanReadable {
		sizeStr = formatHumanReadable(size)
	}

	return fmt.Sprintf("%s %10s %s %s", mode, sizeStr, modTime, name)
}

// formatHumanReadable formats size in human-readable format
func formatHumanReadable(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"K", "M", "G", "T", "P", "E"}
	return fmt.Sprintf("%.1f%s", float64(size)/float64(div), units[exp])
}

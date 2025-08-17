package opt

// Boolean flag types with constants
type LongFormatFlag bool
const (
	LongFormat   LongFormatFlag = true
	ShortFormat  LongFormatFlag = false
)

type AllFilesFlag bool
const (
	AllFiles   AllFilesFlag = true
	NoAllFiles AllFilesFlag = false
)

type HumanReadableFlag bool
const (
	HumanReadable   HumanReadableFlag = true
	NoHumanReadable HumanReadableFlag = false
)

type RecursiveFlag bool
const (
	Recursive   RecursiveFlag = true
	NoRecursive RecursiveFlag = false
)

type ReverseFlag bool
const (
	Reverse   ReverseFlag = true
	NoReverse ReverseFlag = false
)

// Custom types for parameters
type SortBy string

const (
	SortByName SortBy = "name"
	SortByTime SortBy = "time"
	SortBySize SortBy = "size"
)

// Flags represents the configuration options for the ls command
type Flags struct {
	LongFormat    LongFormatFlag    // Long format listing
	AllFiles      AllFilesFlag      // Show all files including hidden
	HumanReadable HumanReadableFlag // Human readable sizes
	Recursive     RecursiveFlag     // Recursive listing
	Reverse       ReverseFlag       // Reverse sort order
	SortBy        SortBy            // Sort criteria
}

// Configure methods for the opt system
func (f LongFormatFlag) Configure(flags *Flags) { flags.LongFormat = f }
func (f AllFilesFlag) Configure(flags *Flags) { flags.AllFiles = f }
func (f HumanReadableFlag) Configure(flags *Flags) { flags.HumanReadable = f }
func (f RecursiveFlag) Configure(flags *Flags) { flags.Recursive = f }
func (f ReverseFlag) Configure(flags *Flags) { flags.Reverse = f }
func (s SortBy) Configure(flags *Flags) { flags.SortBy = s }

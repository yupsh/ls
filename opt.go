package command

type LongFormatFlag bool

const (
	LongFormat  LongFormatFlag = true
	ShortFormat LongFormatFlag = false
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

type SortBy string

const (
	SortByName SortBy = "name"
	SortByTime SortBy = "time"
	SortBySize SortBy = "size"
)

type flags struct {
	LongFormat    LongFormatFlag
	AllFiles      AllFilesFlag
	HumanReadable HumanReadableFlag
	Recursive     RecursiveFlag
	Reverse       ReverseFlag
	SortBy        SortBy
}

func (f LongFormatFlag) Configure(flags *flags)    { flags.LongFormat = f }
func (f AllFilesFlag) Configure(flags *flags)      { flags.AllFiles = f }
func (f HumanReadableFlag) Configure(flags *flags) { flags.HumanReadable = f }
func (f RecursiveFlag) Configure(flags *flags)     { flags.Recursive = f }
func (f ReverseFlag) Configure(flags *flags)       { flags.Reverse = f }
func (s SortBy) Configure(flags *flags)            { flags.SortBy = s }

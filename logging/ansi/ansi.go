package ansi

import "strings"

type EscapeCode int

const (
	Reset EscapeCode = iota
	Bold
	Dim
	Italic
	Underline
	Blink
	Reverse
	Hidden

	// Foreground colors
	FgBlack
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
	FgBrightBlack
	FgBrightRed
	FgBrightGreen
	FgBrightYellow
	FgBrightBlue
	FgBrightMagenta
	FgBrightCyan
	FgBrightWhite

	// Background colors
	BgBlack
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
	BgBrightBlack
	BgBrightRed
	BgBrightGreen
	BgBrightYellow
	BgBrightBlue
	BgBrightMagenta
	BgBrightCyan
	BgBrightWhite
)

var strs = map[EscapeCode]string{
	Reset:     "\033[0m",
	Bold:      "\033[1m",
	Dim:       "\033[2m",
	Italic:    "\033[3m",
	Underline: "\033[4m",
	Blink:     "\033[5m",
	Reverse:   "\033[7m",
	Hidden:    "\033[8m",

	// Foreground colors
	FgBlack:         "\033[30m",
	FgRed:           "\033[31m",
	FgGreen:         "\033[32m",
	FgYellow:        "\033[33m",
	FgBlue:          "\033[34m",
	FgMagenta:       "\033[35m",
	FgCyan:          "\033[36m",
	FgWhite:         "\033[37m",
	FgBrightBlack:   "\033[90m",
	FgBrightRed:     "\033[91m",
	FgBrightGreen:   "\033[92m",
	FgBrightYellow:  "\033[93m",
	FgBrightBlue:    "\033[94m",
	FgBrightMagenta: "\033[95m",
	FgBrightCyan:    "\033[96m",
	FgBrightWhite:   "\033[97m",

	// Background colors
	BgBlack:         "\033[40m",
	BgRed:           "\033[41m",
	BgGreen:         "\033[42m",
	BgYellow:        "\033[43m",
	BgBlue:          "\033[44m",
	BgMagenta:       "\033[45m",
	BgCyan:          "\033[46m",
	BgWhite:         "\033[47m",
	BgBrightBlack:   "\033[100m",
	BgBrightRed:     "\033[101m",
	BgBrightGreen:   "\033[102m",
	BgBrightYellow:  "\033[103m",
	BgBrightBlue:    "\033[104m",
	BgBrightMagenta: "\033[105m",
	BgBrightCyan:    "\033[106m",
	BgBrightWhite:   "\033[107m",
}

type EscapeMode int

const (
	EscapeMode_Enable = iota
	EscapeMode_Disable
)

type AnsiStringBuilder struct {
	str        strings.Builder
	escapeMode EscapeMode
}

func NewAnsiStringBuilder() AnsiStringBuilder {
	return AnsiStringBuilder{escapeMode: EscapeMode_Enable}
}

func (builder *AnsiStringBuilder) SetEscapeMode(mode EscapeMode) {
	builder.escapeMode = mode
}

func (builder *AnsiStringBuilder) WriteString(s string) (int, error) {
	return builder.str.WriteString(s)
}

func (builder *AnsiStringBuilder) WriteEscapeCode(ec EscapeCode) (int, error) {
	if builder.escapeMode == EscapeMode_Disable {
		return 0, nil
	}

	return builder.str.WriteString(strs[ec])
}

func (builder *AnsiStringBuilder) Write(ss ...any) (int, error) {
	n := 0

	for _, s := range ss {
		switch s := s.(type) {
		case EscapeCode:
			nn, err := builder.WriteEscapeCode(s)
			n += nn

			if err != nil {
				return n, err
			}
		case string:
			nn, err := builder.WriteString(s)
			n += nn

			if err != nil {
				return n, err
			}
		}
	}

	return n, nil
}

func (builder *AnsiStringBuilder) String() string {
	return builder.str.String()
}

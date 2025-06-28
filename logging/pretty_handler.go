package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/link00000000/telemetry/logging/ansi"
	"golang.org/x/term"
)

const globalPadding = "                     "

var projectRoot string = "/"

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	projectRoot = filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
}

type PrettyHandler struct {
	writer io.Writer
	level  Level
}

func NewPrettyHandler(writer io.Writer, level Level) PrettyHandler {
	return PrettyHandler{writer: writer, level: level}
}

func (handler PrettyHandler) useColor() bool {
	file, ok := handler.writer.(*os.File)
	if !ok {
		return false
	}

	isTerm := term.IsTerminal(int(file.Fd()))
	return isTerm
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnLoggerCreated(logger *Logger, timestamp time.Time, caller *runtime.Frame) {
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnLoggerClosed(logger *Logger, timestamp time.Time, caller *runtime.Frame) error {
	return nil
}

// Implements [logging.Handler]
func (handler PrettyHandler) HandleRecord(logger *Logger, record Record) error {
	if record.Level < handler.level {
		return nil
	}

	var str ansi.AnsiStringBuilder
	if handler.useColor() {
		str.SetEscapeMode(ansi.EscapeMode_Enable)
	} else {
		str.SetEscapeMode(ansi.EscapeMode_Disable)
	}

	str.Write(record.Time.Format("2006/01/02 15:04:05"), " ")

	switch record.Level {
	case LevelDebug:
		str.Write(ansi.FgMagenta, "DBG", ansi.Reset)
	case LevelInfo:
		str.Write(ansi.FgBlue, "INF", ansi.Reset)
	case LevelWarn:
		str.Write(ansi.FgYellow, "WRN", ansi.Reset)
	case LevelError:
		str.Write(ansi.FgRed, "ERR", ansi.Reset)
	case LevelFatal:
		str.Write(ansi.FgBlack, ansi.BgRed, "FTL", ansi.Reset)
	case LevelPanic:
		str.Write(ansi.FgBlack, ansi.BgRed, "!!!", ansi.Reset)
	}

	str.WriteString(" ")

	var callerRelativePath *string
	if record.Caller != nil {
		if relativePath, err := filepath.Rel(projectRoot, record.Caller.File); err == nil {
			callerRelativePath = &relativePath
		}
	}

	if callerRelativePath != nil {
		str.Write(ansi.FgBrightBlack, fmt.Sprintf("<%s:%d> ", *callerRelativePath, record.Caller.Line), ansi.Reset)
	} else {
		str.Write(ansi.FgBrightBlack, "<UNKNOWN CALLER> ", ansi.Reset)
	}

	str.WriteString(record.Message)

	str.WriteString("\n")

	printAttrsRec(&str, record.Attributes, globalPadding)

	/*
		dataJson, err := json.Marshal(logger.data)
		if err != nil && (strings.Contains(err.Error(), "unsupported type") || strings.Contains(err.Error(), "unsupported value")) {
			// Fallback to non-recursive printing
			printData(&str, logger.data, globalPadding)
		} else if err != nil {
			return err
		} else {
			var dataMap map[string]any
			err = json.Unmarshal([]byte(dataJson), &dataMap)
			if err != nil {
				return err
			}

			printDataRec(&str, dataMap, globalPadding)
		}
	*/

	_, err := fmt.Fprintf(handler.writer, str.String())
	return err
}

func printData(str *ansi.AnsiStringBuilder, data map[string]any, padding string) {
	i := 0
	for k, v := range data {
		str.WriteString(padding)

		isLast := i == len(data)-1
		if !isLast {
			str.WriteString("├─ ")
		} else {
			str.WriteString("└─ ")
		}

		str.Write(ansi.FgBrightBlack, k, ansi.Reset, ": ", fmt.Sprintf("%#v", v), "\n")

		i++
	}
}

func printDataRec(str *ansi.AnsiStringBuilder, data map[string]any, padding string) {
	i := 0
	for k, v := range data {
		str.WriteString(padding)

		isLast := i == len(data)-1
		if !isLast {
			str.WriteString("├─ ")
		} else {
			str.WriteString("└─ ")
		}

		switch v := v.(type) {
		case map[string]any:
			str.Write(ansi.FgBrightBlack, k, ansi.Reset, "\n")

			if !isLast {
				printDataRec(str, v, padding+"│   ")
			} else {
				printDataRec(str, v, padding+"    ")
			}
		default:
			str.Write(ansi.FgBrightBlack, k, ansi.Reset, ": ", fmt.Sprintf("%#v", v), "\n")
		}

		i++
	}
}

func printAttrsRec(str *ansi.AnsiStringBuilder, attrs []Attribute, padding string) {
	for i, attr := range attrs {
		str.WriteString(padding)

		isLast := i == len(attrs)-1
		if !isLast {
			str.WriteString("├─ ")
		} else {
			str.WriteString("└─ ")
		}

		switch v := attr.Value.(type) {
		case []Attribute:
			str.Write(ansi.FgBrightBlack, attr.Key, ansi.Reset, "\n")

			if !isLast {
				printAttrsRec(str, v, padding+"│   ")
			} else {
				printAttrsRec(str, v, padding+"    ")
			}
		case error:
			str.Write(ansi.FgBrightBlack, attr.Key, ansi.Reset, ": ", fmt.Sprintf("%#v \"%s\"", v, v.Error()), "\n")
		default:
			str.Write(ansi.FgBrightBlack, attr.Key, ansi.Reset, ": ", fmt.Sprintf("%#v", v), "\n")
		}
	}
}

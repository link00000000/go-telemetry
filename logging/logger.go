package logging

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

func getModulePath(functionPath string) string {
	// Module paths contain parens for the struct and have an additional dot in the path
	// ex.
	//  Module: github.com/link00000000/fredboard/v3/telemetry.(*Logger).Log
	//  Function: ccidentallycoded.com/fredboard/v3/telemetry.NewLogger
	isMethod := strings.Contains(functionPath, "(")

	var endOfModuleName int
	if isMethod {
		endOfModuleName = strings.LastIndex(functionPath, "(") - 1
	} else {
		endOfModuleName = strings.LastIndex(functionPath, ".")
	}

	if endOfModuleName == -1 {
		panic("malformed module name does not contain a . delimiter")
	}

	return functionPath[:endOfModuleName]
}

var ErrNoCaller = errors.New("no caller")

func getCaller() (*runtime.Frame, error) {
	pcs := make([]uintptr, 8)
	n := runtime.Callers(1, pcs)
	pcs = pcs[:n]

	if len(pcs) == 0 {
		return nil, ErrNoCaller
	}

	frames := runtime.CallersFrames(pcs)

	firstFrame, more := frames.Next()
	if !more {
		return nil, ErrNoCaller
	}

	thisModule := getModulePath(firstFrame.Function)

	for {
		frame, more := frames.Next()
		module := getModulePath(frame.Function)

		if module != thisModule {
			return &frame, nil
		}

		if !more {
			break
		}
	}

	return nil, ErrNoCaller
}

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

type LoggerState int

const (
	LoggerState_Open = iota
	LoggerState_Closed
)

type Record struct {
	Time       time.Time
	Level      Level
	Message    string
	Caller     *runtime.Frame
	Attributes []Attribute
}

type Attribute struct {
	Key   string
	Value any
}

type Handler interface {
	OnLoggerCreated(logger *Logger, time time.Time, caller *runtime.Frame)
	OnLoggerClosed(logger *Logger, time time.Time, caller *runtime.Frame) error
	HandleRecord(logger *Logger, record Record) error
}

type Logger struct {
	id       uuid.UUID
	parent   *Logger
	children []*Logger

	state LoggerState

	panicOnError bool
	handlers     []Handler
}

func NewLogger() *Logger {
	return &Logger{
		id:       uuid.New(),
		children: make([]*Logger, 0),
		state:    LoggerState_Open,
		handlers: make([]Handler, 0),
	}
}

func (logger *Logger) NewChildLogger() *Logger {
	childLogger := NewLogger()
	childLogger.parent = logger

	logger.children = append(logger.children, childLogger)

	caller, err := getCaller()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && err != ErrNoCaller {
		panic(err)
	}

	now := time.Now().UTC()
	for _, handler := range childLogger.Handlers() {
		handler.OnLoggerCreated(childLogger, now, caller)
	}

	return childLogger
}

// Implements [io.Closer]
func (logger *Logger) Close() error {
	// Prevent closing a logger multiple times
	if logger.state == LoggerState_Closed {
		return nil
	}

	errs := make([]error, 0)

	for _, child := range logger.children {
		err := child.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	caller, err := getCaller()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && err != ErrNoCaller {
		return err
	}

	now := time.Now().UTC()
	for _, handler := range logger.Handlers() {
		errs = append(errs, handler.OnLoggerClosed(logger, now, caller))
	}

	logger.state = LoggerState_Closed

	return errors.Join(errs...)
}

func (logger *Logger) RootLogger() *Logger {
	l := logger

	for l.parent != nil {
		l = l.parent
	}

	return l
}

func (logger *Logger) Handlers() []Handler {
	return logger.RootLogger().handlers
}

func (logger *Logger) AddHandler(handler Handler) {
	logger.RootLogger().handlers = append(logger.RootLogger().handlers, handler)
}

func (logger *Logger) PanicOnError() bool {
	return logger.RootLogger().panicOnError
}

func (logger *Logger) SetPanicOnError(value bool) {
	logger.RootLogger().panicOnError = value
}

func (logger *Logger) Log(level Level, message string, args ...any) error {
	caller, err := getCaller()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && !errors.Is(err, ErrNoCaller) {
		return err
	}

	record := Record{
		Time:       time.Now().UTC(),
		Level:      level,
		Message:    message,
		Caller:     caller,
		Attributes: argsToAttrs(args),
	}

	errs := make([]error, 0)
	for _, handler := range logger.Handlers() {
		errs = append(errs, handler.HandleRecord(logger, record))
	}

	return errors.Join(errs...)
}

func (logger *Logger) Debug(message string, args ...any) (err error) {
	err = logger.Log(LevelDebug, message, args...)
	if err != nil && logger.PanicOnError() {
		panic(err)
	}

	return err
}

func (logger *Logger) Info(message string, args ...any) (err error) {
	err = logger.Log(LevelInfo, message, args...)
	if err != nil && logger.PanicOnError() {
		panic(err)
	}

	return err
}

func (logger *Logger) Warn(message string, args ...any) (err error) {
	err = logger.Log(LevelWarn, message, args...)
	if err != nil && logger.PanicOnError() {
		panic(err)
	}

	return err
}

func (logger *Logger) Error(message string, args ...any) (err error) {
	err = logger.Log(LevelError, message, args...)
	if err != nil && logger.PanicOnError() {
		panic(err)
	}

	return err
}

func (logger *Logger) Fatal(message string, args ...any) {
	err := logger.Log(LevelFatal, message, args...)
	if err != nil && logger.PanicOnError() {
		panic(err)
	}

	os.Exit(1)
}

func (logger *Logger) Panic(message string, args ...any) {
	err := logger.Log(LevelPanic, message, args...)
	if err != nil && logger.PanicOnError() {
		panic(err)
	}

	panic("an unrecoverable error has occurred")
}

func argsToAttrs(args []any) (attr []Attribute) {
	remaining := args
	attrs := make([]Attribute, 0)

	for len(remaining) > 0 {
		var attr Attribute
		attr, remaining = nextAttrFromArgs(remaining)
		attrs = append(attrs, attr)
	}

	return attrs
}

func nextAttrFromArgs(args []any) (attr Attribute, remaining []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return Attribute{Key: "!BADKEY", Value: x}, nil
		}
		return Attribute{Key: x, Value: args[1]}, args[2:]
	default:
		return Attribute{Key: "!BADKEY", Value: x}, args[1:]
	}
}

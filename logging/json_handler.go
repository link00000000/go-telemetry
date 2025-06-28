package logging

import (
	"encoding/json"
	"io"
	"runtime"
	"time"
)

type JsonHandlerMessageType int

const (
	JsonHandlerMessageType_LoggerCreated JsonHandlerMessageType = iota
	JsonHandlerMessageType_LoggerClosed
	JsonHandlerMessageType_Record
)

type JsonHandlerCaller struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type JsonHandlerLogger struct {
	Id       string   `json:"id"`
	Parent   *string  `json:"parent"`
	Children []string `json:"children"`
	Root     string   `json:"root"`
}

type JsonHandlerLoggerCreated struct {
	Time   time.Time         `json:"time"`
	Caller JsonHandlerCaller `json:"caller"`
	Logger JsonHandlerLogger `json:"logger"`
}

type JsonHandlerLoggerClosed struct {
	Time   time.Time         `json:"time"`
	Caller JsonHandlerCaller `json:"caller"`
	Logger JsonHandlerLogger `json:"logger"`
}

type JsonHandlerRecord struct {
	Time    time.Time         `json:"time"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Error   *string           `json:"error"`
	Caller  JsonHandlerCaller `json:"caller"`
	Logger  JsonHandlerLogger `json:"logger"`
}

type JsonHandlerMessage[T any] struct {
	Type JsonHandlerMessageType `json:"type"`
	Data T                      `json:"data"`
}

func NewJsonLoggerCreatedMessage() JsonHandlerMessage[JsonHandlerLoggerCreated] {
	return JsonHandlerMessage[JsonHandlerLoggerCreated]{Type: JsonHandlerMessageType_LoggerCreated, Data: JsonHandlerLoggerCreated{}}
}

func NewJsonLoggerClosedMessage() JsonHandlerMessage[JsonHandlerLoggerClosed] {
	return JsonHandlerMessage[JsonHandlerLoggerClosed]{Type: JsonHandlerMessageType_LoggerClosed, Data: JsonHandlerLoggerClosed{}}
}

func NewJsonLoggerRecordMessage() JsonHandlerMessage[JsonHandlerRecord] {
	return JsonHandlerMessage[JsonHandlerRecord]{Type: JsonHandlerMessageType_Record, Data: JsonHandlerRecord{}}
}

type JsonHandler struct {
	writer io.Writer
	level  Level
}

func NewJsonHandler(writer io.Writer, level Level) JsonHandler {
	return JsonHandler{writer: writer, level: level}
}

// Implements [logging.Handler]
func (handler JsonHandler) OnLoggerCreated(logger *Logger, timestamp time.Time, caller *runtime.Frame) {
	loggerCreated := NewJsonLoggerCreatedMessage()
	loggerCreated.Data.Time = timestamp

	loggerCreated.Data.Caller = JsonHandlerCaller{}
	loggerCreated.Data.Caller.File = caller.File
	loggerCreated.Data.Caller.Line = caller.Line

	loggerCreated.Data.Logger.Id = logger.id.String()
	loggerCreated.Data.Logger.Root = logger.RootLogger().id.String()

	if logger.parent != nil {
		str := logger.parent.id.String()
		loggerCreated.Data.Logger.Parent = &str
	}

	loggerCreated.Data.Logger.Children = make([]string, len(logger.children))
	for i, c := range logger.children {
		loggerCreated.Data.Logger.Children[i] = c.id.String()
	}

	data, err := json.Marshal(loggerCreated)
	if err != nil {
		panic(err)
	}

	// TODO: Handle error?
	handler.writer.Write(append(data, byte('\n')))
}

// Implements [logging.Handler]
func (handler JsonHandler) OnLoggerClosed(logger *Logger, timestamp time.Time, caller *runtime.Frame) error {
	loggerClosed := NewJsonLoggerClosedMessage()
	loggerClosed.Data.Time = timestamp

	loggerClosed.Data.Caller = JsonHandlerCaller{}
	loggerClosed.Data.Caller.File = caller.File
	loggerClosed.Data.Caller.Line = caller.Line

	loggerClosed.Data.Logger.Id = logger.id.String()
	loggerClosed.Data.Logger.Root = logger.RootLogger().id.String()

	if logger.parent != nil {
		str := logger.parent.id.String()
		loggerClosed.Data.Logger.Parent = &str
	}

	loggerClosed.Data.Logger.Children = make([]string, len(logger.children))
	for i, c := range logger.children {
		loggerClosed.Data.Logger.Children[i] = c.id.String()
	}

	data, err := json.Marshal(loggerClosed)
	if err != nil {
		return err
	}

	handler.writer.Write(append(data, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}

// Implements [logging.Handler]
func (handler JsonHandler) HandleRecord(logger *Logger, record Record) error {
	if record.Level < handler.level {
		return nil
	}

	message := NewJsonLoggerRecordMessage()
	message.Data.Time = record.Time

	switch record.Level {
	case LevelDebug:
		message.Data.Level = "debug"
	case LevelInfo:
		message.Data.Level = "info"
	case LevelWarn:
		message.Data.Level = "warn"
	case LevelError:
		message.Data.Level = "error"
	case LevelFatal:
		message.Data.Level = "fatal"
	case LevelPanic:
		message.Data.Level = "panic"
	}

	message.Data.Message = record.Message

	message.Data.Caller = JsonHandlerCaller{}
	message.Data.Caller.File = record.Caller.File
	message.Data.Caller.Line = record.Caller.Line

	message.Data.Logger.Id = logger.id.String()
	message.Data.Logger.Root = logger.RootLogger().id.String()

	if logger.parent != nil {
		str := logger.parent.id.String()
		message.Data.Logger.Parent = &str
	}

	message.Data.Logger.Children = make([]string, len(logger.children))
	for i, c := range logger.children {
		message.Data.Logger.Children[i] = c.id.String()
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	handler.writer.Write(append(data, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}

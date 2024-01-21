package net

import (
	"context"
	"errors"
	"io"
	slog "log/slog"
	"os"
	"sync"

	logz "log/slog"

	"github.com/hedzr/is/basics"
)

type baseS struct {
	logger        Logger
	loggerHandler slog.Handler
	closers       Closers
}

func newBaseS() baseS {
	// return baseS{logger: logz.New("net-enh").WithSkip(1)}
	return baseS{logger: newDefaultLogger()}
}

type Closable interface {
	Close()
}

type Closers []Closable

func (s Closers) Close() {
	for _, c := range s {
		c.Close()
	}
}

type cw struct {
	obj io.Closer
}

func (s *cw) Close() {
	if err := s.obj.Close(); err != nil {
		println("obj.Close() failed.", err, s.obj)
	}
}

type cf struct {
	fn func()
}

func (s *cf) Close() {
	if s.fn != nil {
		s.fn()
	}
}

func (s *baseS) Close() {
	s.closers.Close()
	s.closers = nil

	if c, ok := s.logger.(interface{ Close() }); ok {
		c.Close()
	}

	basics.Close() // call to hedzr/is/basics/closers.Close
}

func (s *baseS) addClosable(closable Closable) { s.closers = append(s.closers, closable) }
func (s *baseS) addCloser(closer io.Closer)    { s.closers = append(s.closers, &cw{closer}) }
func (s *baseS) addCloseFunc(fn func())        { s.closers = append(s.closers, &cf{fn}) }

func (s *baseS) Verbose(msg string, args ...any) {
	// s.logger.Verbose(msg, args...)
	// s.logger.Info(msg, args...)
	s.logger.Log(context.Background(), LevelVerbose, msg, args...)
}
func (s *baseS) Trace(msg string, args ...any) {
	// s.logger.Trace(msg, args...)
	// s.logger.Info(msg, args...)
	s.logger.Log(context.Background(), LevelTrace, msg, args...)
}
func (s *baseS) Debug(msg string, args ...any) {
	// s.logger.Debug(msg, args...)
	s.logger.Debug(msg, args...)
}
func (s *baseS) Fatal(msg string, args ...any) {
	// s.logger.Panic(msg, args...)
	// s.logger.Error(msg, args...)
	s.logger.Log(context.Background(), LevelFatal, msg, args...)
	panic(msg)
}
func (s *baseS) Panic(msg string, args ...any) {
	s.logger.Log(context.Background(), LevelPanic, msg, args...)
	panic(msg)
}
func (s *baseS) Error(msg string, args ...any) { s.logger.Error(msg, args...) }
func (s *baseS) Warn(msg string, args ...any)  { s.logger.Warn(msg, args...) }
func (s *baseS) Info(msg string, args ...any)  { s.logger.Info(msg, args...) }
func (s *baseS) Hint(msg string, args ...any) {
	// s.logger.Trace(msg, args...)
	// s.logger.Info(msg, args...)
	s.logger.Log(context.Background(), LevelHint, msg, args...)
}
func (s *baseS) Notice(msg string, args ...any) {
	s.logger.Log(context.Background(), LevelNotice, msg, args...)
}
func (s *baseS) Print(msg string, args ...any) {
	// s.logger.Print(msg, args...)
	s.logger.Log(context.Background(), LevelHint, msg, args...)
}
func (s *baseS) Println(args ...any) {
	// s.logger.Println(args...)
	s.logger.Log(context.Background(), LevelHint, args[0].(string), args[1:]...)
}
func (s *baseS) Log(ctx context.Context, lvl slog.Level, msg string, args ...any) {
	s.logger.Log(ctx, lvl, msg, args...)
}
func (s *baseS) Logger() Logger        { return s.logger }
func (s *baseS) DefaultLogger() Logger { return newDefaultLogger() }

func (s *baseS) setLoggerHandler(h slog.Handler) {
	s.loggerHandler = h

	if h == nil {
		s.logger = newDefaultLogger()
	} else {
		s.logger = slog.New(h)
	}
}

func (s *baseS) handleError(err error, reason string, args ...any) {
	if err != nil {
		if len(reason) > 0 && reason[0] == '[' {
			s.logger.Error(reason, append([]any{"err", err}, args...)...)
		} else {
			s.logger.Error("ERROR", append([]any{"err", err, "reason", reason}, args...)...)
		}
	}
}

//

//

//

func (s *baseS) DefaultHandlerOptions() slog.HandlerOptions {
	return newDefaultHandlerOptions()
}

func newDefaultHandlerOptions() slog.HandlerOptions {
	opts := slog.HandlerOptions{
		Level: LevelTrace,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := LevelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	}
	return opts
}

func newDefaultLogger() Logger {
	onceLogger.Do(func() {
		opts := newDefaultHandlerOptions()
		defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	})
	return defaultLogger
}

var defaultLogger Logger

var onceLogger sync.Once

const (
	LevelVerbose = slog.Level(-16)
	LevelTrace   = slog.Level(-8)
	LevelNotice  = slog.Level(2)
	LevelHint    = slog.Level(3)
	LevelFatal   = slog.Level(16)
	LevelPanic   = slog.Level(17)
)

var LevelNames = map[slog.Leveler]string{
	LevelVerbose: "VERBOSE",
	LevelTrace:   "TRACE",
	LevelNotice:  "NOTICE",
	LevelHint:    "HINT",
	LevelFatal:   "FATAL",
	LevelPanic:   "PANIC",
}

//

//

//

type Logger interface {
	LogEntry
}

type LogEntry interface {
	// Close()

	// String() string
	// Level() Level

	BasicLogger
}

type BasicLogger interface {
	Printer
	// Enabled(requestingLevel Level) bool // to test the requesting logging level should be allowed.
	// EnabledContext(ctx context.Context, requestingLevel Level) bool

	// WithSkip create a new child logger with specified extra
	// ignored stack frames, which will be plussed over the
	// internal stack frames stripping tool.
	//
	// A child logger is super lite commonly. It'll take a little
	// more resource usages only if you have LattrsR set globally.
	// In that case, child logger looks up all its parents for
	// collecting all attributes and logging them.
	// WithSkip(extraFrames int) Entry
}

type Printer interface {
	Error(msg string, args ...any) // error
	Warn(msg string, args ...any)  // warning
	Info(msg string, args ...any)  // info. Attr, Attrs in args will be recognized as is
	Debug(msg string, args ...any) // only for state.Env().InDebugging() or IsDebugBuild()

	Log(ctx context.Context, lvl logz.Level, msg string, args ...any)
}

type ExtraPrinters interface {
	// Close()

	Verbose(msg string, args ...any) // only for -tags=verbose
	Print(msg string, args ...any)   // logging always
	Println(args ...any)             // synonym to Print, NOTE first elem of args decoded as msg here

	Trace(msg string, args ...any) // only for state.Env().InTracing()
	Panic(msg string, args ...any) // error and panic
	Fatal(msg string, args ...any) // error and os.Exit(-3)

	// OK(msg string, args ...any)      // identify it is in OK mode
	// Success(msg string, args ...any) // identify a successful operation done
	// Fail(msg string, args ...any)    // identify a wrong occurs, default to stderr device
}

var errMethodNotAllowed = errors.New("method not allowed")

var errUnimplemented = errors.New("unimplemented")

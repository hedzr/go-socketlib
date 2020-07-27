/*
 * Copyright © 2020 Hedzr Yeh.
 */

package logger_old

import (
	"context"
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/trace"
	"github.com/sirupsen/logrus"
)

type Base struct {
	Tag string
	logrus.Fields
	// logrus.Entry
}

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *logrus.Entry {
	return logrus.WithError(err)
}

// WithContext creates an entry from the standard logger and adds a context to it.
func WithContext(ctx context.Context) *logrus.Entry {
	return logrus.WithContext(ctx)
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}

//

func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Tracef(format string, args ...interface{}) {
	if !silent() {
		logrus.Tracef(format, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if !silent() {
		logrus.Debugf(format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if !silent() {
		logrus.Infof(format, args...)
	}
}

//
// -------------------------
//

func (s *Base) Wrong(e error, fmt string, args ...interface{}) {
	s.checkFields().WithError(e).Errorf(fmt, args...)
}

// With can add key-value pair to logger context and print them later.
//
// ~~With could cause race exception, for example: when two or more
// clients incoming with CONNECT packets, now connectParser parse them
// ok and print logging info via base.With("sth", sb).Debug(...), here
// is the race point.~~
//
// NOTE: data race had solved.
//
// Another scene is the multiple clients send pingreq packets.
func (s *Base) With(key string, value interface{}) *Base {
	newBase := &Base{Tag: s.Tag, Fields: make(logrus.Fields)}
	newBase.Fields[key] = value
	return newBase
}

func (s *Base) With2(key string, value interface{}, key2 string, value2 interface{}) *Base {
	newBase := &Base{Tag: s.Tag, Fields: make(logrus.Fields)}
	newBase.Fields[key] = value
	newBase.Fields[key2] = value2
	return newBase
}

func (s *Base) With3(key string, value interface{}, key2 string, value2 interface{}, key3 string, value3 interface{}) *Base {
	newBase := &Base{Tag: s.Tag, Fields: make(logrus.Fields)}
	newBase.Fields[key] = value
	newBase.Fields[key2] = value2
	newBase.Fields[key3] = value3
	return newBase
}

func (s *Base) Skip(level int) *Base {
	newBase := &Base{Tag: s.Tag, Fields: make(logrus.Fields)}
	newBase.Fields[logex.SKIP] = level
	return newBase
}

func (s *Base) checkFields() *logrus.Entry {
	var fields logrus.Fields = s.Fields
	if fields == nil {
		fields = make(logrus.Fields)
	}

	if _, ok := fields["C"]; !ok {
		fields["C"] = s.Tag
	}
	if _, ok := fields[logex.SKIP]; !ok {
		fields[logex.SKIP] = 1
	}
	return logrus.WithFields(fields)
}

func (s *Base) Debug(fmt string, args ...interface{}) {
	if !silent() {
		s.checkFields().Debugf(fmt, args...)
	}
}

func (s *Base) Printf(fmt string, args ...interface{}) {
	s.checkFields().Printf(fmt, args...)
}

func (s *Base) Print(args ...interface{}) {
	s.checkFields().Print(args...)
}

func (s *Base) Info(fmt string, args ...interface{}) {
	if !silent() {
		s.checkFields().Infof(fmt, args...)
	}
}

func (s *Base) Warn(fmt string, args ...interface{}) {
	s.checkFields().Warnf(fmt, args...)
}

func (s *Base) Trace(fmt string, args ...interface{}) {
	if trace.IsEnabled() && !silent() {
		s.checkFields().Printf(fmt, args...)
	}
}

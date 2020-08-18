/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A fabricLogger is an adapter around a zap.SugaredLogger that provides
// structured logging capabilities while preserving much of the legacy logging
// behavior.
//
// The most significant difference between the fabricLogger and the
// zap.SugaredLogger is that methods without a formatting suffix (f or w) build
// the log entry message with fmt.Sprintln instead of fmt.Sprint. Without this
// change, arguments are not separated by spaces.
type fabricLogger struct{ s *zap.SugaredLogger }

func (f *fabricLogger) DPanic(args ...interface{})                    { f.s.DPanicf(formatArgs(args)) }
func (f *fabricLogger) DPanicf(template string, args ...interface{})  { f.s.DPanicf(template, args...) }
func (f *fabricLogger) DPanicw(msg string, kvPairs ...interface{})    { f.s.DPanicw(msg, kvPairs...) }
func (f *fabricLogger) Debug(args ...interface{})                     { f.s.Debugf(formatArgs(args)) }
func (f *fabricLogger) Debugf(template string, args ...interface{})   { f.s.Debugf(template, args...) }
func (f *fabricLogger) Debugw(msg string, kvPairs ...interface{})     { f.s.Debugw(msg, kvPairs...) }
func (f *fabricLogger) Error(args ...interface{})                     { f.s.Errorf(formatArgs(args)) }
func (f *fabricLogger) Errorf(template string, args ...interface{})   { f.s.Errorf(template, args...) }
func (f *fabricLogger) Errorw(msg string, kvPairs ...interface{})     { f.s.Errorw(msg, kvPairs...) }
func (f *fabricLogger) Fatal(args ...interface{})                     { f.s.Fatalf(formatArgs(args)) }
func (f *fabricLogger) Fatalf(template string, args ...interface{})   { f.s.Fatalf(template, args...) }
func (f *fabricLogger) Fatalw(msg string, kvPairs ...interface{})     { f.s.Fatalw(msg, kvPairs...) }
func (f *fabricLogger) Info(args ...interface{})                      { f.s.Infof(formatArgs(args)) }
func (f *fabricLogger) Infof(template string, args ...interface{})    { f.s.Infof(template, args...) }
func (f *fabricLogger) Infow(msg string, kvPairs ...interface{})      { f.s.Infow(msg, kvPairs...) }
func (f *fabricLogger) Panic(args ...interface{})                     { f.s.Panicf(formatArgs(args)) }
func (f *fabricLogger) Panicf(template string, args ...interface{})   { f.s.Panicf(template, args...) }
func (f *fabricLogger) Panicw(msg string, kvPairs ...interface{})     { f.s.Panicw(msg, kvPairs...) }
func (f *fabricLogger) Warn(args ...interface{})                      { f.s.Warnf(formatArgs(args)) }
func (f *fabricLogger) Warnf(template string, args ...interface{})    { f.s.Warnf(template, args...) }
func (f *fabricLogger) Warnw(msg string, kvPairs ...interface{})      { f.s.Warnw(msg, kvPairs...) }
func (f *fabricLogger) Warning(args ...interface{})                   { f.s.Warnf(formatArgs(args)) }
func (f *fabricLogger) Warningf(template string, args ...interface{}) { f.s.Warnf(template, args...) }

// for backwards compatibility
func (f *fabricLogger) Critical(args ...interface{})                   { f.s.Errorf(formatArgs(args)) }
func (f *fabricLogger) Criticalf(template string, args ...interface{}) { f.s.Errorf(template, args...) }
func (f *fabricLogger) Notice(args ...interface{})                     { f.s.Infof(formatArgs(args)) }
func (f *fabricLogger) Noticef(template string, args ...interface{})   { f.s.Infof(template, args...) }

func (f *fabricLogger) Named(name string) *fabricLogger { return &fabricLogger{s: f.s.Named(name)} }
func (f *fabricLogger) Sync() error                     { return f.s.Sync() }
func (f *fabricLogger) Zap() *zap.Logger                { return f.s.Desugar() }

func (f *fabricLogger) IsEnabledFor(level zapcore.Level) bool {
	return f.s.Desugar().Core().Enabled(level)
}

func (f *fabricLogger) With(args ...interface{}) *fabricLogger {
	return &fabricLogger{s: f.s.With(args...)}
}

func (f *fabricLogger) WithOptions(opts ...zap.Option) *fabricLogger {
	l := f.s.Desugar().WithOptions(opts...)
	return &fabricLogger{s: l.Sugar()}
}

func formatArgs(args []interface{}) string { return strings.TrimSuffix(fmt.Sprintln(args...), "\n") }

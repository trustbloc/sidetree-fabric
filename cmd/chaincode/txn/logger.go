/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"fmt"
)

// Logger is a simple implementation of a logger that uses fmt.Printf
// TODO: This should be replaced by a real logger
type Logger struct {
	module string
}

// NewLogger returns a new logger
func NewLogger(module string) *Logger {
	return &Logger{
		module: module,
	}
}

// Debugf prints a log
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.printf("DEBUG", format, args...)
}

// Infof prints a log
func (l *Logger) Infof(format string, args ...interface{}) {
	l.printf("INFO", format, args...)
}

// Warnf prints a log
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.printf("WARN", format, args...)
}

// Errorf prints a log
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.printf("ERROR", format, args...)
}

func (l *Logger) printf(level, format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("[%s] -> %s %s\n", l.module, level, format), args...)
}

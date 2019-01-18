/*
	Package for structured logging with levels: Fatal, Error, Info, Debug.

	Default is Error level with only Fatal and Error messages logged.
	Use Debug level for development to log all the messages.
	Use Info level for end users that want to have extra info in the logs.

	New(w io.WriteCloser) is for opening logging to custom writer.

	log.Default.Level = Debug // Switch to Debug level.
	log.Default.Time = false // Switch off time information for systemd.

	Copyright (C) 2018 Etasoft Inc.
*/
package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Type of a logging level.
type Level int

// Standard logging levels.
const (
	FatalLevel Level = iota
	ErrorLevel
	InfoLevel
	DebugLevel
)

var levelToLetter = map[Level]string{
	FatalLevel: "F",
	ErrorLevel: "E",
	InfoLevel:  "_",
	DebugLevel: ".",
}

// A Logger represents logging type that used writer.
type Logger struct {
	// Minimum level to log.
	Level Level
	Time  bool

	w io.WriteCloser
	sync.Mutex
}

// New logger based on io.WriteCloser.
func New(w io.WriteCloser) *Logger {
	return &Logger{
		w:     w,
		Level: ErrorLevel,
		Time:  true,
	}
}

// Close the writer behind the logger.
func (l *Logger) Close() {
	l.w.Close()
}

// log the message into the logger, at the given level.
// skip is the number of frames to skip when computing the file name and line number.
func (l *Logger) log(level Level, skip int, format string, a ...interface{}) {
	if level > l.Level {
		return
	}

	// Message.
	msg := fmt.Sprintf(format, a...)

	// Caller.
	_, file, line, ok := runtime.Caller(2 + skip)
	if !ok {
		file = "unknown"
	}
	fl := fmt.Sprintf("%s:%-4d", filepath.Base(file), line)
	if len(fl) > 18 {
		fl = fl[len(fl)-18:]
	}
	msg = fmt.Sprintf("%-18s", fl) + " " + msg

	// Level.
	letter, ok := levelToLetter[level]
	if !ok {
		letter = strconv.Itoa(int(level))
	}
	msg = letter + " " + msg

	// Time.
	if l.Time {
		msg = time.Now().Format("2006/01/02 15:04:05 ") + msg
	}

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	l.Lock()
	l.w.Write([]byte(msg))
	l.Unlock()
}

// Debugf logs information at a Debug level.
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.log(DebugLevel, 1, format, a...)
}

// Infof logs information at a Info level.
func (l *Logger) Infof(format string, a ...interface{}) {
	l.log(InfoLevel, 1, format, a...)
}

// Errorf logs information at an Error level. It also returns an error
// constructed with the given message, in case it's useful for the caller.
func (l *Logger) Errorf(format string, a ...interface{}) error {
	l.log(ErrorLevel, 1, format, a...)
	return fmt.Errorf(format, a...)
}

// Error logs information at an Error level.
func (l *Logger) Error(err error) error {
	l.log(ErrorLevel, 1, "Error: %v\n", err)
	return err
}

// Fatalf logs information at a Fatal level, and then exits the program with a
// non-0 exit code.
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.log(FatalLevel, 1, format, a...)
	os.Exit(1)
}

// Fatal logs information at a Fatal level, and then exits the program with a
// non-0 exit code.
func (l *Logger) Fatal(err error) {
	l.log(FatalLevel, 1, "Error: %v\n", err)
	os.Exit(1)
}

// The default logger, used by the top-level functions below.
var Default = &Logger{
	w:     os.Stderr,
	Level: ErrorLevel,
	Time:  true,
}

// Debugf is a convenient wrapper to Default.Debugf.
func Debugf(format string, a ...interface{}) {
	Default.Debugf(format, a...)
}

// Infof is a convenient wrapper to Default.Infof.
func Infof(format string, a ...interface{}) {
	Default.Infof(format, a...)
}

// Errorf is a convenient wrapper to Default.Errorf.
func Errorf(format string, a ...interface{}) error {
	return Default.Errorf(format, a...)
}

func Error(err error) error {
	return Default.Error(err)
}

// Fatalf is a convenient wrapper to Default.Fatalf.
func Fatalf(format string, a ...interface{}) {
	Default.Fatalf(format, a...)
}

// Fatal is a convenient wrapper to Default.Fatal.
func Fatal(err error) {
	Default.Fatal(err)
}

package flog

import (
	"fmt"
	"os"
	"time"
)

type Level int

const None Level = -1
const (
	Debug Level = iota
	Info
	Warn
	Error
	Fatal
)

var (
	minLevel = Info
	logCh    = make(chan string, 1024)
)

func init() {

}

func SetLevel(l int) {
	minLevel = Level(l)
	if l != -1 {
		go func() {
			for msg := range logCh {
				fmt.Fprint(os.Stdout, msg)
			}
		}()
	}
}

func logf(level Level, format string, args ...any) {
	if level < minLevel || minLevel == None {
		return
	}

	for _, arg := range args {
		if err, ok := arg.(error); ok {
			err = WErr(err)
			if err == nil {
				return
			}
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("%s [%s] %s\n", now, level.String(), fmt.Sprintf(format, args...))

	select {
	case logCh <- line:
	default:
	}
}

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	case None:
		return "None"
	default:
		return "UNKNOWN"
	}
}

func Debugf(format string, args ...any) { logf(Debug, format, args...) }
func Infof(format string, args ...any)  { logf(Info, format, args...) }
func Warnf(format string, args ...any)  { logf(Warn, format, args...) }
func Errorf(format string, args ...any) { logf(Error, format, args...) }
func Fatalf(format string, args ...any) {
	// For fatal errors, we must ensure the message is delivered
	// Use blocking write instead of select with default
	if minLevel != None && Fatal >= minLevel {
		// Check if any errors should suppress logging
		// This matches the behavior in logf()
		for _, arg := range args {
			if err, ok := arg.(error); ok {
				err = WErr(err)
				if err == nil {
					// Non-critical error, exit without logging
					os.Exit(1)
				}
			}
		}

		now := time.Now().Format("2006-01-02 15:04:05.000")
		line := fmt.Sprintf("%s [%s] %s\n", now, Fatal.String(), fmt.Sprintf(format, args...))
		
		// Blocking write to ensure fatal message is always sent
		// This is the key fix - use blocking write instead of select with default
		logCh <- line
		// Give the logger goroutine time to flush
		time.Sleep(50 * time.Millisecond)
	}
	os.Exit(1)
}

func Close() { close(logCh) }

package logger

import (
	"fmt"
	"path"
	"runtime"
	"strconv"
	"time"
)

var simpleLogger = SimpleLogger{LogLevelDebug}

type Logger interface {
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	DebugfNoCR(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}
type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

type SimpleLogger struct {
	level LogLevel
}

func (sl *SimpleLogger) GetPrefix(level string) string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(4, fpcs)
	caller := ""
	if n != 0 {
		caller1 := runtime.FuncForPC(fpcs[0] - 1)
		_, caller2 := path.Split(caller1.Name())

		file, lineNo := caller1.FileLine(fpcs[0] - 1)
		if caller1 != nil {
			_, fileName := path.Split(file)
			caller = caller2 + " " + fileName + "(" + strconv.Itoa(lineNo) + ")"
		}
	}

	return time.Now().Format("Jan 2 15:04:05.000") + " " + level + " " + caller + ":"
}

func (sl *SimpleLogger) Trace(v ...interface{}) {
	if sl.level <= LogLevelTrace {
		arr := []interface{}{sl.GetPrefix("[Trace]")}
		for _, item := range v {
			arr = append(arr, item)
		}

		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) Tracef(format string, v ...interface{}) {
	if sl.level <= LogLevelTrace {
		fmt.Printf(sl.GetPrefix("[Trace]")+format+"\n", v...)
	}
}

func (sl *SimpleLogger) Debug(v ...interface{}) {
	if sl.level <= LogLevelDebug {
		arr := []interface{}{sl.GetPrefix("[Debug]")}
		for _, item := range v {
			arr = append(arr, item)
		}

		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) Debugf(format string, v ...interface{}) {
	if sl.level <= LogLevelDebug {
		fmt.Printf(sl.GetPrefix("[Debug]")+format+"\n", v...)
	}
}
func (sl *SimpleLogger) Info(v ...interface{}) {
	if sl.level <= LogLevelInfo {
		arr := []interface{}{sl.GetPrefix("[Info ]")}
		for _, item := range v {
			arr = append(arr, item)
		}
		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) DebugfNoCR(format string, v ...interface{}) {
	if sl.level <= LogLevelDebug {
		fmt.Printf(sl.GetPrefix("[Info ]")+format, v...)
	}
}

func (sl *SimpleLogger) Infof(format string, v ...interface{}) {
	if sl.level <= LogLevelInfo {
		fmt.Printf(sl.GetPrefix("[Info ]")+format+"\n", v...)
	}
}

func (sl *SimpleLogger) Warn(v ...interface{}) {
	if sl.level <= LogLevelWarn {
		arr := []interface{}{sl.GetPrefix("[Warn ]")}
		for _, item := range v {
			arr = append(arr, item)
		}
		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) Warnf(format string, v ...interface{}) {
	if sl.level <= LogLevelWarn {
		fmt.Printf(sl.GetPrefix("[Warn ]")+format+"\n", v...)
	}
}

func (sl *SimpleLogger) Error(v ...interface{}) {
	if sl.level <= LogLevelError {
		arr := []interface{}{sl.GetPrefix("[Error]")}
		for _, item := range v {
			arr = append(arr, item)
		}
		fmt.Println(arr...)
	}
}

func (sl *SimpleLogger) Errorf(format string, v ...interface{}) {
	if sl.level <= LogLevelError {
		fmt.Printf(sl.GetPrefix("[Error]")+format+"\n", v...)
	}
}

func (sl *SimpleLogger) Fatal(v ...interface{}) {
	if sl.level <= LogLevelFatal {
		arr := []interface{}{sl.GetPrefix("[Fatal]")}
		for _, item := range v {
			arr = append(arr, item)
		}
		fmt.Println(arr...)

	}
}

func (sl *SimpleLogger) Fatalf(format string, v ...interface{}) {
	if sl.level <= LogLevelFatal {
		fmt.Printf(sl.GetPrefix("[Fatal]")+format+"\n", v)
	}
}

func Trace(v ...interface{}) {
	simpleLogger.Trace(v...)
}

func Tracef(format string, v ...interface{}) {
	simpleLogger.Tracef(format, v...)
}

func Debug(v ...interface{}) {
	simpleLogger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	simpleLogger.Debugf(format, v...)
}

func Info(v ...interface{}) {
	simpleLogger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	simpleLogger.Infof(format, v...)
}

func DebugfNoCR(format string, v ...interface{}) {
	simpleLogger.DebugfNoCR(format, v...)
}

func Warn(v ...interface{}) {
	simpleLogger.Warn(v...)
}
func Warnf(format string, v ...interface{}) {
	simpleLogger.Warnf(format, v...)
}

func Error(v ...interface{}) {
	simpleLogger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	simpleLogger.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	simpleLogger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	simpleLogger.Fatalf(format, v...)
}

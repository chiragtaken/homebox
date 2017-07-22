/*
 *    Copyright (c) 2017 by Cisco Systems, Inc.
 *    All rights reserved.
 *    AUTHOR : Suvil Deora (sudeora@cisco.com)
 */

package AfwCommon

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	AFW_LOGGER_LEVEL_FATAL   uint16 = 0
	AFW_LOGGER_LEVEL_ERROR   uint16 = 1
	AFW_LOGGER_LEVEL_WARNING uint16 = 2
	AFW_LOGGER_LEVEL_INFO    uint16 = 3
	AFW_LOGGER_LEVEL_DEBUG   uint16 = 4
	AFW_LOG_FILE_SIZE        int64  = 10 * 1024 * 1024
)

type LogLine struct {
	LogLine string
	Level   uint16
}

type Logger struct {
	LogCh       chan LogLine
	LogDebug    *log.Logger
	LogInfo     *log.Logger
	LogWarning  *log.Logger
	LogError    *log.Logger
	LogFatal    *log.Logger
	logLevel    uint16
	file        *os.File
	fileName    string
	currVersion int
}

func (l *Logger) Init(fileName string) {

	var err error
	var versionFile string

	l.LogCh = make(chan LogLine)
	l.fileName = fileName
	l.logLevel = AFW_LOGGER_LEVEL_ERROR

	versionFile = fileName + ".1"
	l.currVersion = 1
	os.Remove(l.fileName)
	os.Remove(versionFile)

	l.file, err = os.OpenFile(versionFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", ":", err)
	}

	l.LogDebug = log.New(l.file, "DEBUG: ", log.Ldate|log.Ltime)

	l.LogInfo = log.New(l.file, "INFO: ", log.Ldate|log.Ltime)

	l.LogWarning = log.New(l.file, "WARNING: ", log.Ldate|log.Ltime)

	l.LogError = log.New(l.file, "ERROR: ", log.Ldate|log.Ltime)

	l.LogFatal = log.New(l.file, "FATAL: ", log.Ldate|log.Ltime) //|log.Lshortfile)

	err = os.Symlink(versionFile, l.fileName)
	if err != nil {
		fmt.Println("Unable to create symbolic link to the log version file", err)
	}

	go func() {
		for {
			line := <-l.LogCh
			switch line.Level {
			case AFW_LOGGER_LEVEL_DEBUG:
				l.LogDebug.Println(line.LogLine)
			case AFW_LOGGER_LEVEL_INFO:
				l.LogInfo.Println(line.LogLine)
			case AFW_LOGGER_LEVEL_WARNING:
				l.LogWarning.Println(line.LogLine)
			case AFW_LOGGER_LEVEL_ERROR:
				l.LogError.Println(line.LogLine)
			case AFW_LOGGER_LEVEL_FATAL:
				l.LogFatal.Println(line.LogLine)
			default:
				l.LogDebug.Println(line.LogLine)
			}
			l.CheckSize()
		}
	}()
}

func (l *Logger) SetLoglevel(level string) {
	var levelInt uint16
	switch level {
	case "DEBUG":
		levelInt = AFW_LOGGER_LEVEL_DEBUG
	case "INFO":
		levelInt = AFW_LOGGER_LEVEL_INFO
	case "ERROR":
		levelInt = AFW_LOGGER_LEVEL_ERROR
	case "WARNING":
		levelInt = AFW_LOGGER_LEVEL_WARNING
	case "FATAL":
		levelInt = AFW_LOGGER_LEVEL_FATAL
	default:
		levelInt = AFW_LOGGER_LEVEL_FATAL
	}
	l.logLevel = levelInt
}

func (l *Logger) GetCaller() (string, int) {

	_, f, n, _ := runtime.Caller(2)

	x := strings.Split(f, "/")
	return x[len(x)-1], n
}

func (l *Logger) Fatal(args ...interface{}) {

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintln(fileName, lineno, args)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_FATAL}
}

func (l *Logger) Fatalf(format string, args ...interface{}) {

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintf("%s:%d"+format, fileName, lineno, args)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_FATAL}

}

func (l *Logger) Errorf(format string, args ...interface{}) {

	if l.logLevel < AFW_LOGGER_LEVEL_ERROR {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintf("%s:%d "+format, fileName, lineno, args)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_ERROR}
}

func (l *Logger) Error(errStr ...interface{}) {

	if l.logLevel < AFW_LOGGER_LEVEL_ERROR {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintln(fileName, lineno, errStr)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_ERROR}
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	if l.logLevel < AFW_LOGGER_LEVEL_WARNING {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintf("%s:%d "+format, fileName, lineno, args)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_WARNING}
}

func (l *Logger) Warning(warningStr ...interface{}) {
	if l.logLevel < AFW_LOGGER_LEVEL_WARNING {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintln(fileName, lineno, warningStr)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_WARNING}

}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.logLevel < AFW_LOGGER_LEVEL_INFO {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintf("%s:%d "+format, fileName, lineno, args)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_INFO}
}

func (l *Logger) Info(infoStr ...interface{}) {
	if l.logLevel < AFW_LOGGER_LEVEL_INFO {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintln(fileName, lineno, infoStr)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_INFO}

}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.logLevel < AFW_LOGGER_LEVEL_DEBUG {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintf("%s:%d "+format, fileName, lineno, args)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_DEBUG}
}

func (l *Logger) Debug(debugStr ...interface{}) {
	if l.logLevel < AFW_LOGGER_LEVEL_DEBUG {
		return
	}

	fileName, lineno := l.GetCaller()
	line := fmt.Sprintln(fileName, lineno, debugStr)
	l.LogCh <- LogLine{line, AFW_LOGGER_LEVEL_DEBUG}

}

func (l *Logger) gotoNextVersion() {

	var err error
	var f *os.File

	if l.currVersion == 10 {
		l.currVersion = 1
	} else {
		l.currVersion++
	}

	oldFile := l.file
	versionFile := l.fileName + "." + strconv.Itoa(l.currVersion)
	os.Remove(versionFile)
	f, err = os.OpenFile(versionFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		l.LogError.Println("Failed to open log file", ":", err)
		return
	}

	l.LogFatal.SetOutput(f)
	l.LogDebug.SetOutput(f)
	l.LogInfo.SetOutput(f)
	l.LogWarning.SetOutput(f)
	l.LogError.SetOutput(f)

	os.Remove(l.fileName)
	os.Symlink(versionFile, l.fileName)
	oldFile.Close()
	l.file = f
}

func (l *Logger) CheckSize() {
	fileInfo, err := l.file.Stat()
	if err != nil {
		fmt.Println("Error checking file size", err)
		return
	}
	if fileInfo.Size() > AFW_LOG_FILE_SIZE {
		l.gotoNextVersion()
	}
}

package utils

import (
	"bufio"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// TODO: log file rotate

// TODO: add opentracing

// Init init logger config
func Init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

func SetLogLevel(level string) {
	var logLevel log.Level
	switch level {
	case "TRACE": logLevel = log.TraceLevel
	case "DEBUG": logLevel = log.DebugLevel
	case "INFO": logLevel = log.InfoLevel
	case "WARN": logLevel = log.WarnLevel
	case "ERROR": logLevel = log.ErrorLevel
	case "FATAL": logLevel = log.FatalLevel
	case "PANIC": logLevel = log.PanicLevel
	default:
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
}

func fileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func AddFileOutputToLog(filepath string) {
	baseDir := path.Dir(filepath)
	if !fileOrDirExists(baseDir) {
		_ = os.MkdirAll(baseDir, os.ModePerm)
	}
	f, err := os.Create(filepath)
	if err != nil {
		println(err.Error())
		panic("can't open log file " + filepath)
		return
	}
	err = f.Sync()
	if err != nil {
		println(err.Error())
		panic(err.Error())
		return
	}
	fileWriter := bufio.NewWriter(f)
	log.SetOutput(fileWriter)
}

// GetLogger get logger instance of module
func GetLogger(module string) log.FieldLogger {
	return log.WithFields(log.Fields{
		"module": module,
	})
}

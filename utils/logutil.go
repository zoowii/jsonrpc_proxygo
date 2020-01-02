package utils

import (
	"log"
)

// TODO: add logger and opentracing

const Dev = true

func Debug(a ...interface{}) {
	if !Dev {
		return
	}
	log.Println(a...)
}

func Debugf(format string, v ...interface{}) {
	if !Dev {
		return
	}
	log.Printf(format, v...)
}

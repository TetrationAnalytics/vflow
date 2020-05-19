package log

import (
	"io/ioutil"
	"log"
)

var (
	// Logger is the instance of a StdLogger interface that vFlow writes connection
	// management events to. By default it is set to discard all log messages via ioutil.Discard,
	// but you can set it to redirect wherever you want.
	Logger StdLogger = log.New(ioutil.Discard, "[vflow] ", log.LstdFlags)
)

// StdLogger is used to log error messages.
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

package log

import (
	"log"
	"os"
	"io/ioutil"
)

// Logs
var Err *log.Logger
var Out *log.Logger
var Debug *log.Logger
var Report *log.Logger


func InitLoggers(debugFlag bool) {
	Err = log.New(os.Stderr, "!!! ", log.LstdFlags)
	Out = log.New(os.Stdout, "    ", log.LstdFlags)
	if debugFlag {
		Debug = log.New(os.Stdout, "(d) ", log.LstdFlags)
	} else {
		Debug = log.New(ioutil.Discard, "(d) ", log.LstdFlags)
	}
	Report = log.New(os.Stdout, "+++ ", log.LstdFlags)
}

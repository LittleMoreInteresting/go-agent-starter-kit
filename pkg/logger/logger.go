package logger

import (
	"log"
	"os"
)

func New() *log.Logger {
	return log.New(os.Stdout, "[agent-kit] ", log.LstdFlags|log.Lshortfile)
}

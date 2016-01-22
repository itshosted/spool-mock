package config

import (
	"log"
	"os"
)

var (
	Verbose bool
	L *log.Logger
	RequeMsgids []string
)

func Init() error {
	L = log.New(os.Stdout, "spool-mock ", log.LstdFlags)
	return nil
}
package config

import (
	"log"
	"os"
)

var (
	Verbose bool
	L *log.Logger
)

func Init() error {
	L = log.New(os.Stdout, "spool-mock ", log.LstdFlags)
	return nil
}
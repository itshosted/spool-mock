package main

import (
	"flag"
	"spool-mock/config"
)

func main() {
	nntp := ""

	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&nntp, "n", "0.0.0.0:9091", "NNTP Listen on ip:port")
	flag.Parse()

	if e := nntpListen(nntp); e != nil {
		panic(e)
	}
}
package main

import (
	"flag"
	"fmt"
	"github.com/p4ali/httpgo/httpgo"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s -port 8000 -name httpgo\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

var version = "undefined"
var port = 8000
var name = "httpgo"

func main() {
	flag.Usage = usage

	if len(os.Args) == 2 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Fprintf(os.Stdout, "%s\n", version)
			os.Exit(0)
		}
	}

	portPtr := flag.Int("port", port, "Server port (default 8000)")
	namePtr := flag.String("name", name, "Server name")
	flag.Parse()

	httpgo.NewServer(*namePtr, version, true).Start(httpgo.GetIP(), *portPtr, httpgo.GetHostname())
}

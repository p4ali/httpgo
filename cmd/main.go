package main

import (
	"flag"
	"fmt"
	"github.com/p4ali/httpgo/httpgo"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s -port 8000 -name httpgo -version 0.0.1\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usage

	portPtr := flag.Int("port", -1, "Server port (Required)")
	namePtr := flag.String("name", "httpgo", "Server name")
	versionPtr := flag.String("version", "0.0.1", "Server version")
	flag.Parse()

	// guard required parameters
	if *portPtr == -1 {
		fmt.Fprintf(os.Stderr, "Missing required port\n")
		flag.Usage()
	}

	httpgo.NewServer(*namePtr, *versionPtr).Start(*portPtr)
}

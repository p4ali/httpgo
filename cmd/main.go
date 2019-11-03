// Package main is the main entry, it handles arguments
package main

import (
	"flag"
	"fmt"
	"github.com/p4ali/httpgo/httpgo"
	"net"
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

// GetIP return the ip address of the host
func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				os.Stdout.WriteString(ipnet.IP.String() + "\n")
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

// GetHostname return hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return hostname
}

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

	httpgo.NewServer(*namePtr, version, true).Start(getIP(), *portPtr, getHostname())
}

package httpgo

import (
	"net"
	"net/http"
	"os"
	"strings"
)

func GetEnvs() map[string]string {
	envs := make(map[string]string)
	for _, s := range os.Environ() {
		kv := strings.SplitN(s, "=", 2)
		envs[kv[0]] = kv[1]
	}
	return envs
}

// GetIP return the ip address of the host
func GetIP() string {
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

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return hostname
}

func Chain(x http.HandlerFunc, y http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x(w, r)
		y(w, r)
	}
}

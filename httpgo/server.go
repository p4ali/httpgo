// Package httpgo provide REST Server and APIs
package httpgo

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Server status
type Server struct {
	Name       string      `json:"name,omitempty"`      // server name
	Version    string      `json:"version,omitempty"`   // server version
	Port       int         `json:"port,omitempty"`      // port to listen
	Healthy    bool        `json:"healthy,omitempty"`   // server healthy or not. If not all calls except /health POST will return 503
	Host       string      `json:"hostname,omitempty"`  // hostname
	IP         string      `json:"ip,omitempty"`        // host IP
	Router     *mux.Router `json:"-"`                   // router manage the routes and handlers
	Client     http.Client `json:"-"`                   // http client to call other APIs
	HTTPServer http.Server `json:"-"`                   // underline http server
}

// NewServer create a http server, given the name, version and health status of the server
func NewServer(name string, version string, healthy bool) *Server {
	return (&Server{
		Name:    name,
		Version: version,
		Healthy: healthy,
		Router:  mux.NewRouter(),
	}).route()
}

// Start a http server
func (s *Server) Start(ip string, port int, host string) {
	s.IP = ip
	s.Host = host
	s.Healthy = true
	s.Port = port
	s.Client = http.Client{Timeout: time.Duration(int(60)) * time.Second}
	fmt.Println("listening to ", s.Port)
	svr := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: s.Router}
	log.Print(svr.ListenAndServe())
}

//////////// Routing table ///////////////
func (s *Server) route() *Server {
	s.Router.HandleFunc("/callother", s.injectHeaderThen(s.handleCallOther())).Methods("POST")
	s.Router.HandleFunc("/debug", s.injectHeaderThen(s.handleDebug())).Methods("GET")
	s.Router.HandleFunc("/delay/{ms}", s.injectHeaderThen(s.handleDelay())).Methods("GET")
	s.Router.HandleFunc("/echo/{msg}", s.injectHeaderThen(s.handleEcho())).Methods("GET")
	s.Router.HandleFunc("/health", s.injectHeaderThen(s.handleHealth())).Queries("value", "{true|false}").Methods("POST")
	s.Router.HandleFunc("/health", s.injectHeaderThen(s.handleHealth())).Methods("GET")
	s.Router.HandleFunc("/health", s.injectHeaderThen(s.handleHealth())).Methods("HEAD")
	s.Router.HandleFunc("/name", s.injectHeaderThen(s.handleName())).Methods("GET")
	s.Router.HandleFunc("/name", s.injectHeaderThen(s.handleName())).Queries("value", "{.*}").Methods("POST")
	s.Router.HandleFunc("/status/{code}", s.injectHeaderThen(s.handleStatus())).Methods("GET")
	return s
}

//////////// handlers ///////////////
func (s *Server) handleCallOther() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error while attempting to read body of incoming request.")
			w.WriteHeader(400)
			return
		}
		urls := strings.Split(string(body[:]), "\r\n")
		var doSync = false
		if val, exist := r.URL.Query()["sync"]; exist {
			doSync = strings.ToUpper(val[0]) == "TRUE"
			log.Println("Will do synchronous invocations per request override value.")
		}
		c := make(chan string)
		for _, url := range urls {
			go s.invokeURL(url, r, c)
			if doSync {
				w.Write([]byte(<-c))
			}
		}
		if !doSync {
			for range urls {
				w.Write([]byte(<-c))
			}
		}
	}
}

func (s *Server) invokeURL(url string, orig *http.Request, c chan string) {
	urlResp := "<<Error>>"
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range orig.Header {
		req.Header.Set(k, v[0])
		log.Println("set header " + k + ": " + v[0])
	}
	log.Println("Invoking url: ", url)
	resp, urlErr := s.Client.Do(req)
	if urlErr == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			log.Println("Response :", resp.Status)
			log.Println("Response body :", string(body[:]))
			urlResp = url + ": " + string(body[:]) + "\n"
		} else {
			log.Println("Error during invocation, status-code=", resp.StatusCode)
		}
	} else {
		log.Println("Error while trying to prepare for url request: ", urlErr)
	}
	c <- urlResp
}

func (s *Server) handleDebug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(s.getDebugInfo())
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			return
		}
	}
}

func (s *Server) handleDelay() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		i64, err := strconv.ParseInt(params["ms"], 10, 64)
		if err != nil {
			log.Print(err)
			w.WriteHeader(400)
			return
		}
		time.Sleep(time.Duration(int(i64)) * time.Millisecond)
		w.Write([]byte(fmt.Sprintf("Delayed %d", int(i64))))
	}
}

func (s *Server) handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		i64, err := strconv.ParseInt(params["code"], 10, 64)
		if err != nil {
			log.Print(err)
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(int(i64))
	}
}

func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "HEAD":
			fallthrough
		case "GET":
			if !s.Healthy {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			break
		case "POST":
			set := r.URL.Query().Get("value")
			c := http.StatusOK
			healthy, err := strconv.ParseBool(set)
			if err != nil {
				log.Print(err)
				c = http.StatusBadRequest
			} else {
				s.Healthy = healthy
			}
			w.WriteHeader(c)
			break
		}
	}
}

func (s *Server) handleEcho() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		w.Write([]byte(params["msg"]))
	}
}

func (s *Server) handleName() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			c := http.StatusOK
			w.WriteHeader(c)
			w.Write([]byte(s.Name))
			break
		case "POST":
			s.Name = r.URL.Query().Get("value")
			c := http.StatusOK
			w.WriteHeader(c)
			break
		}
	}
}

/////////////// Util struct and methods ///////////////////////////
// debugResponse represents debug info in each response
type debugResponse struct {
	Server Server            `json:"server,omitempty"`
	Env    map[string]string `json:"environments,omitempty"`
}

func (s *Server) getDebugInfo() *debugResponse {
	info := &debugResponse{}
	info.Env = getEnvs()
	return info
}

// getEnvs get environment variable as string
func getEnvs() map[string]string {
	envs := make(map[string]string)
	for _, s := range os.Environ() {
		kv := strings.SplitN(s, "=", 2)
		envs[kv[0]] = kv[1]
	}
	return envs
}

// inject header before calling the handler x
func (s *Server) injectHeaderThen(x http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeHeader()(w, r)
		x(w, r)
	}
}

// write server state as header for debug
func (s *Server) writeHeader() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Response request to " + r.URL.String() + " by client (" + r.RemoteAddr + ")")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// allow pre-flight headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Range, Content-Disposition, Content-Type, ETag")
		w.Header().Set("echo-server-ip", s.IP)
		w.Header().Set("echo-server-version", s.Version)
		w.Header().Set("echo-server-name", s.Name)
		if !s.Healthy {
			if r.Method == http.MethodPost && r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}
	}
}

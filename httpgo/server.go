// Package httpgo provide REST Server and APIs
package httpgo

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Server status
type Server struct {
	Name    string      `json:"name,omitempty"`
	Version string      `json:"version,omitempty"`
	Port    int         `json:"port,omitempty"`
	Healthy bool        `json:"healthy,omitempty"`
	Host    string      `json:"hostname,omitempty"`
	IP      string      `json:"ip,omitempty"`
	Router  *mux.Router `json:"-"`
	Client  http.Client `json:"-"`
}

// DebugResponse represents debug info in each response
type DebugResponse struct {
	Server Server            `json:"server,omitempty"`
	Env    map[string]string `json:"environments,omitempty"`
}

// NewServer create a http server
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
	log.Print(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Router))
}

//////////// Routing table and handlers ///////////////
func (s *Server) route() *Server {
	s.Router.HandleFunc("/callother", s.c(s.handleCallOther())).Methods("POST")
	s.Router.HandleFunc("/debug", s.c(s.handleDebug())).Methods("GET")
	s.Router.HandleFunc("/delay/{ms}", s.c(s.handleDelay())).Methods("GET")
	s.Router.HandleFunc("/echo/{msg}", s.c(s.handleEcho())).Methods("GET")
	s.Router.HandleFunc("/health", s.c(s.handleHealth())).Queries("value", "{true|false}").Methods("POST")
	s.Router.HandleFunc("/health", s.c(s.handleHealth())).Methods("GET")
	s.Router.HandleFunc("/health", s.c(s.handleHealth())).Methods("HEAD")
	s.Router.HandleFunc("/name", s.c(s.handleName())).Methods("GET")
	s.Router.HandleFunc("/name", s.c(s.handleName())).Queries("name", "{.*}").Methods("POST")
	s.Router.HandleFunc("/status/{code}", s.c(s.handleStatus())).Methods("GET")
	return s
}

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
		json.NewEncoder(w).Encode(s.getDebugInfo())
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
			s.Name = r.URL.Query().Get("name")
			c := http.StatusOK
			w.WriteHeader(c)
			break
		}
	}
}

/////////////// Private methods ///////////////////////////
func (s *Server) getDebugInfo() *DebugResponse {
	info := &DebugResponse{Server: *s}
	info.Env = GetEnvs()
	return info
}

func (s *Server) c(x http.HandlerFunc) http.HandlerFunc {
	return Chain(s.writeHeader(), x)
}

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

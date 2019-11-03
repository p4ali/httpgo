package httpgo_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/p4ali/httpgo/httpgo"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// To start a new Server on host of "turing" with nick name "foo", version "1.1.0" and healthy "true":
//   import "github.com/p4ali/httpgo/httpgo"
func Example() {
	httpgo.NewServer("foo", "1.1.0", true).Start("127.0.0.1", 8000, "turing")
}

var server *httptest.Server
var client = &http.Client{}

func TestMain(m *testing.M) {
	handler := httpgo.NewServer("test", "0.0.1", true)
	server = httptest.NewServer(handler.Router)
	defer server.Close()
	code := m.Run()
	os.Exit(code)
}

func TestHttpGet(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Health check should return 200",
			method:         "GET",
			url:            fmt.Sprintf("%s/health", server.URL),
			expectedStatus: 200,
		},
		{
			name:           "Echo should return same message",
			method:         "GET",
			url:            fmt.Sprintf("%s/echo/hi", server.URL),
			expectedStatus: 200,
			expectedBody:   "hi",
		},
	}
	for _, test := range tests {
		log.Print(fmt.Sprintf("TEST: %s %s", test.method, test.url))
		req, err := http.NewRequest(test.method, test.url, nil)
		checkError(t, err)
		resp, err := client.Do(req)
		checkError(t, err)
		checkResponseCode(t, test.expectedStatus, resp.StatusCode)
		if resp.StatusCode != test.expectedStatus {
			t.Fatalf("%s: %s response status %d != %d\n", test.method, test.url, resp.StatusCode, test.expectedStatus)
		}

		matchBody(t, test.expectedBody, resp)
		log.Print("PASS")
	}
}

func TestHealth(t *testing.T) {
	urls := []string{
		fmt.Sprintf("%s/health?", server.URL),
		fmt.Sprintf("%s/debug?", server.URL),
		fmt.Sprintf("%s/echo/x?", server.URL),
		fmt.Sprintf("%s/delay/123?", server.URL),
	}

	// negative
	resp, err := http.Post(fmt.Sprintf("%s/health?value=false", server.URL), "application/json", nil)
	checkError(t, err)
	checkResponseCode(t, 200, resp.StatusCode)

	for _, url := range urls {
		resp, err = http.Get(url)
		checkError(t, err)
		checkResponseCode(t, http.StatusServiceUnavailable, resp.StatusCode)
	}
	// positive
	resp, err = http.Post(fmt.Sprintf("%s/health?value=true", server.URL), "application/json", nil)
	checkError(t, err)
	checkResponseCode(t, 200, resp.StatusCode)

	for _, url := range urls {
		resp, err = http.Get(url)
		checkError(t, err)
		checkResponseCode(t, http.StatusOK, resp.StatusCode)
	}

}

func TestGetStatus(t *testing.T) {
	codes := []int{200, 400, 500}
	for _, code := range codes {
		resp, err := http.Get(fmt.Sprintf("%s/status/%d", server.URL, code))
		checkError(t, err)
		checkResponseCode(t, code, resp.StatusCode)
	}
}

func TestGetDelay(t *testing.T) {
	delay := int64(100)
	start := time.Now()
	resp, err := http.Get(fmt.Sprintf("%s/delay/%d", server.URL, delay))
	checkError(t, err)
	checkResponseCode(t, http.StatusOK, resp.StatusCode)
	sec := time.Since(start).Milliseconds()

	t.Log(fmt.Sprintf("Expected delay %d, Got %d\n", delay, sec))
	if sec < delay || sec > int64(delay*2) {
		t.Errorf("Expected delay %d, Got %d\n", delay, sec)
	}
}

func TestGetDebug(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/debug", server.URL))
	checkError(t, err)
	checkResponseCode(t, http.StatusOK, resp.StatusCode)
	matchBody(t, ".*server.*name.*", resp)
}

func TestPostCallOther(t *testing.T) {
	body := strings.Join([]string{"http://www.google.com"}, "\r\n")
	// async
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/callother", server.URL), bytes.NewBuffer([]byte(body)))
	checkError(t, err)
	req.Header.Set("Accept-Encoding", "*")
	resp, err := client.Do(req)
	checkError(t, err)
	checkResponseCode(t, 200, resp.StatusCode)
	matchBody(t, ".*google.com*", resp)

	// sync
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/callother?sync=true", server.URL), bytes.NewBuffer([]byte(body)))
	checkError(t, err)
	req.Header.Set("Accept-Encoding", "*")
	resp, err = client.Do(req)
	checkError(t, err)
	checkResponseCode(t, 200, resp.StatusCode)
	matchBody(t, ".*google.com*", resp)
}

func TestPostName(t *testing.T) {
	// set
	resp, err := http.Post(fmt.Sprintf("%s/name?value=foo", server.URL), "application/json", nil)
	checkError(t, err)
	checkResponseCode(t, 200, resp.StatusCode)

	// get
	resp, err = http.Get(fmt.Sprintf("%s/name", server.URL))
	checkError(t, err)
	checkResponseCode(t, 200, resp.StatusCode)
	matchBody(t, "foo", resp)
}

func TestStart(t *testing.T) {
	server := httpgo.NewServer("test", "0.0.1", true)
	go func() {
		server.Start("127.0.0.1", 12345, "localhost")
	}()
	defer server.HTTPServer.Close()
	task := func() (bool, error) {
		url := "http://127.0.0.1:12345/debug"
		resp, err := http.Get(url)
		if err == nil {
			if 200 != resp.StatusCode {
				err = fmt.Errorf("Expected response code %d. Got %d", 200, resp.StatusCode)
			}else{
				log.Print("Got 200 calling HTTPServer ", url)
			}
		}
		return err == nil, err
	}
	_, err := timeout(60*time.Second, 2*time.Second, task)
	checkError(t, err)
}

// Helper functions
func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func matchBody(t *testing.T, expectedBody string, resp *http.Response) {
	if expectedBody != "" {
		respBody, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		body := string(respBody)
		match, err := regexp.MatchString(expectedBody, body)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Fatalf("Expected response body %s. Got %s\n", expectedBody, body)
		}
	}
}

func timeout(total time.Duration, interval time.Duration, doSomething func()(bool, error)) (bool, error) {
	timeout := time.After(total)
	tick := time.Tick(interval)
	for {
		select {
		case <-timeout:
			return false, errors.New("timed out")
		case <-tick:
			ok, err := doSomething()
			if err != nil {
				log.Print("doSomething failed, retry... ",err)
			} else if ok {
				log.Print("doSomething succeed!")
				return true, nil
			}
		}
	}
}
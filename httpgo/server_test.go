package httpgo

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	handler := NewServer("test", "0.0.1", true)
	server = httptest.NewServer(handler.Router)
	defer server.Close()
	code := m.Run()
	os.Exit(code)
}

func TestHttpgo(t *testing.T) {

	client := &http.Client{}

	tests := []struct {
		name string
		method string
		url string
		expectedStatus int
		expectedBody string
	}{
		{
			name: "Health check should return 200",
			method: "GET",
			url:  fmt.Sprintf("%s/health", server.URL),
			expectedStatus: 200,
		},
		{
			name: "Echo should return same message",
			method: "GET",
			url:  fmt.Sprintf("%s/echo/hi", server.URL),
			expectedStatus: 200,
			expectedBody: "hi",
		},
	}
	for _, test := range tests {
		req, err := http.NewRequest(test.method, test.url, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != test.expectedStatus {
			t.Fatalf("%s: %s response status %d != %d\n", test.method, test.url, resp.StatusCode, test.expectedStatus)
		}
		log.Print(fmt.Sprintf("PASS: GET %s", test.url))

		if test.expectedBody != "" {
			func(){
				respBody, _ := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()
				body := string(respBody)
				t.Log(body)
				if test.expectedBody != body {
					t.Fatalf("%s: %s response body %s != %s\n", test.method, test.url, body, test.expectedBody)
				}
			}()
		}
	}
}
package httpgo

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpgo(t *testing.T) {
	handler := NewServer("test", "0.0.1")
	server := httptest.NewServer(handler)
	defer server.Close()

	for _, i := range []int{1, 2} {
		log.Print(i)
		resp, err := http.Get(fmt.Sprintf("%s/health", server.URL))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("GET: Received non-200 response: iter[%d] status=%d\n", i, resp.StatusCode)
		}
		log.Print(fmt.Sprintf("PASS: GET %s/health", server.URL))

		resp, err = http.Head(fmt.Sprintf("%s/health", server.URL))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("HEAD: Received non-200 response: iter[%d] status=%d\n", i, resp.StatusCode)
		}
		log.Print(fmt.Sprintf("PASS: HEAD %s/health", server.URL))
	}
}

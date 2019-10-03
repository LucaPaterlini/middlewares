package logger

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

type FakeResponse struct {
	t       *testing.T
	headers http.Header
	body    []byte
	status  int
}

func (r *FakeResponse) Header() http.Header {
	return r.headers
}

func (r *FakeResponse) Write(body []byte) (int, error) {
	r.body = body
	return len(body), nil
}

func (r *FakeResponse) WriteHeader(status int) {
	r.status = status
}

func FakeResponseNew(t *testing.T) *FakeResponse {
	return &FakeResponse{
		t:       t,
		headers: make(http.Header),
	}
}

func TestLogRequest(t *testing.T) {
	// checking the log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	req, _ := http.NewRequest("GET", "hello/test", nil)
	handlerToTest := LogRequest(http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	handlerToTest.ServeHTTP(FakeResponseNew(t), req)
	if !strings.HasSuffix(buf.String(), "GET hello/test\n") {
		t.Errorf("Expected: %s\n Got: %s", "GET hello/test\n", buf.String())
	}
}

func TestLogRequestPanic(t *testing.T) {
	// checking the log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	req, _ := http.NewRequest("GET", "hello/test", nil)
	handlerToTest := LogRequestPanic(http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("hello I like to panic")
	})))
	handlerToTest.ServeHTTP(FakeResponseNew(t), req)
	if !strings.HasSuffix(buf.String(), "Panic Recovered: hello I like to panic\n") {
		t.Errorf("Expected: %s\n Got: %s", "Panic Recovered: hello I like to panic\n", buf.String())
	}
}

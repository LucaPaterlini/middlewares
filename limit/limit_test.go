package limit

import (
	"github.com/go-test/deep"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVisitors_addVisitorIP(t *testing.T) {
	limit := Visitors{
		CleanupRefreshTime: time.Second,
		CleanupExpiry:      2 * time.Second,
		R:                  2,
		B:                  3,
	}
	now := time.Now()
	limit.register = make(map[string]*visitor)
	limit.addVisitorIP("1.2.3.4", now)
	// verify
	if diffList := deep.Equal(*limit.register["1.2.3.4"], visitor{rate.NewLimiter(limit.R, limit.B), now}); len(diffList) > 0 {
		t.Errorf("Diff    : %v\n", diffList)
	}
}

func TestVisitors_getVisitorIP(t *testing.T) {
	limit := Visitors{
		CleanupRefreshTime: time.Second,
		CleanupExpiry:      2 * time.Second,
		R:                  2,
		B:                  3,
	}
	now := time.Now()
	limit.register = make(map[string]*visitor)
	limit.addVisitorIP("1.2.3.4", now)
	limit.addVisitorIP("1.2.3.4", now)
	// verify
	if diffList := deep.Equal(*limit.getVisitorIP("1.2.3.4"), *rate.NewLimiter(limit.R, limit.B)); len(diffList) > 0 {
		t.Errorf("Diff    : %v\n", diffList)
	}
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func TestVisitors_cleanupVisitors(t *testing.T) {
	limit := Visitors{
		CleanupRefreshTime: time.Second,
		CleanupExpiry:      2 * time.Second,
		R:                  2,
		B:                  3,
	}
	limit.register = make(map[string]*visitor)
	limit.addVisitorIP("hello", time.Now())
	limit.cleanupVisitors()
	// than wait that the cleaner can go in execution
	time.Sleep(3 * time.Second)
}

func TestVisitors_Limit(t *testing.T) {
	// load test
	limit := Visitors{
		CleanupRefreshTime: time.Second,
		CleanupExpiry:      3 * time.Second,
		R:                  2,
		B:                  3,
	}
	ts := httptest.NewServer(limit.Limit(http.HandlerFunc(okHandler), true))
	defer ts.Close()

	// set the client

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req.Header.Set("X-Real-IP", "1.2.3.4")

	testExpectedCode := []int{200, 200, 200, 429}

	// test load multiple ports
	for i := 0; i < 4; i++ {
		res, err := client.Do(req)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		if testExpectedCode[i] != res.StatusCode {
			t.Errorf("Expected: %d, got : %d", testExpectedCode[i], res.StatusCode)
		}
	}
}

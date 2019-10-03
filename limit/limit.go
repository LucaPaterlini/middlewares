package limit

import (
	"net/http"
	"sync"
	"time"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Visitors contain the list of visitors of this instance of the website.
type Visitors struct {
	register           map[string]*visitor
	mtx                sync.RWMutex
	CleanupRefreshTime time.Duration
	CleanupExpiry      time.Duration
	R                  rate.Limit
	B                  int
}

func (v *Visitors) addVisitorIP(ip string, time time.Time) *rate.Limiter {
	limiter := rate.NewLimiter(v.R, v.B)
	v.mtx.Lock()
	v.register[ip] = &visitor{limiter, time}
	v.mtx.Unlock()
	return limiter
}

func (v *Visitors) getVisitorIP(addr string) *rate.Limiter {
	v.mtx.Lock()
	item, exists := v.register[addr]
	if !exists {
		v.mtx.Unlock()
		return v.addVisitorIP(addr, time.Now())
	}
	item.lastSeen = time.Now()
	v.mtx.Unlock()
	return item.limiter
}

func (v *Visitors) cleanupVisitors() {
	ticker := time.NewTicker(v.CleanupRefreshTime)
	go func() {
		for range ticker.C {
			v.mtx.Lock()
			for ip, visitor := range v.register {
				if time.Now().Add(-v.CleanupExpiry).After(visitor.lastSeen) {
					delete(v.register, ip)
				}
			}
			v.mtx.Unlock()
		}
	}()
}

// Limit works as a limiter for a specific function handler, added a parameter to deactivate the control.
func (v *Visitors) Limit(next http.Handler, active bool) http.Handler {
	// initialization
	go v.cleanupVisitors()
	v.register = make(map[string]*visitor)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if active {
			limiter := v.getVisitorIP(r.Header.Get("X-Real-IP"))
			if !limiter.Allow() {
				http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

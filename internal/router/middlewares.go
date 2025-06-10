package router

import (
	"fmt"
	"net/http"

	"github.com/BrunoPolaski/go-logger/logger"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(1, 5)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			fmt.Sprintf(
				"----------------- Received request ----------------- \nMethod: %s\nPath: %s\nHeaders: %v\n",
				r.Method,
				r.URL.Path,
				r.Header,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			logger.Warn("Rate limit exceeded")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

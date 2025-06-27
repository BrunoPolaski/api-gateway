package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BrunoPolaski/go-logger/logger"
	"github.com/BrunoPolaski/go-rest-err/rest_err"
	"golang.org/x/time/rate"
)

type httpRequest struct {
	w    http.ResponseWriter
	r    *http.Request
	next http.Handler
}

var limiter = rate.NewLimiter(1, 5)
var requestQueue = make(chan httpRequest, 100)

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
	go processRequest()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		req := httpRequest{w, r, next}

		select {
		case requestQueue <- req:
		default:
			logger.Warn("Server queue full")

			restErr := rest_err.NewRestErr(
				"Server busy, try again later",
				"Request buffer full",
				http.StatusServiceUnavailable,
				[]rest_err.Causes{},
			)

			w.WriteHeader(restErr.Code)
			encoder.Encode(restErr)
			return
		}
	})
}

func processRequest() {
	for req := range requestQueue {
		_ = limiter.Wait(req.r.Context())
		req.next.ServeHTTP(req.w, req.r)
	}
}

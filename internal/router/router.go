package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/BrunoPolaski/api-gateway/internal/entities"
	"github.com/BrunoPolaski/go-logger/logger"
	"github.com/BrunoPolaski/go-rest-err/rest_err"
)

func Init() *http.ServeMux {
	r := http.NewServeMux()

	r.Handle(
		"/",
		LoggingMiddleware(
			RateLimitMiddleware(
				http.HandlerFunc(handleGateway),
			),
		),
	)

	r.Handle("/health", LoggingMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				logger.Error("Failed to write response: " + err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			logger.Info("Health check successful")
		})))

	return r
}

func handleGateway(w http.ResponseWriter, r *http.Request) {
	var routes []entities.Route

	fileBytes, err := os.ReadFile("internal/router/route_table.json")
	if err != nil {
		logger.Error("Failed to read route table: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(fileBytes, &routes)
	if err != nil {
		logger.Error("Failed to parse route table: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, route := range routes {
		if (route.Method == "*" || route.Method == r.Method) && strings.HasPrefix(r.URL.Path, route.Path) {
			logger.Info(fmt.Sprintf("Forwarding request to %s %s%s", r.Method, route.Target, r.URL.Path))
			forwardRequest(w, r, route.Target)
			return
		}
	}
	http.NotFound(w, r)
}

func forwardRequest(w http.ResponseWriter, r *http.Request, target string) {
	targetUrl, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.Director = func(request *http.Request) {
		request.URL.Scheme = targetUrl.Scheme
		request.URL.Host = targetUrl.Host
		request.URL.Path = targetUrl.Path
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		restErr := rest_err.NewRestErr(
			"Bad Gateway",
			err.Error(),
			http.StatusBadGateway,
			nil,
		)

		http.Error(w, restErr.Err, restErr.Code)
	}
	proxy.ServeHTTP(w, r)
}

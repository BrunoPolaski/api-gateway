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

var routes []entities.Route

func Init() *http.ServeMux {
	r := http.NewServeMux()

	fileBytes, err := os.ReadFile("internal/router/route_table.json")
	if err != nil {
		logger.Error("Failed to read route table: " + err.Error())
		return nil
	}
	err = json.Unmarshal(fileBytes, &routes)
	if err != nil {
		logger.Error("Failed to parse route table: " + err.Error())
		return nil
	}

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
	for _, route := range routes {
		if (route.Method == "*" || strings.EqualFold(route.Method, r.Method)) && strings.HasPrefix(r.URL.Path, route.Path) {
			logger.Info(fmt.Sprintf("Forwarding request to %s %s%s", r.Method, route.Target, r.URL.Path))
			forwardRequest(w, r, route.Target, route.Path)
			return
		}
	}
	http.NotFound(w, r)
}

func forwardRequest(w http.ResponseWriter, r *http.Request, target string, path string) {
	targetUrl, err := url.Parse(target)
	if err != nil {
		logger.Error("Failed to parse target URL: " + err.Error())
		restErr := rest_err.NewRestErr(
			"Bad Gateway",
			"Invalid target URL",
			http.StatusBadGateway,
			nil,
		)
		http.Error(w, restErr.Err, restErr.Code)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.Director = func(request *http.Request) {
		request.URL.Scheme = targetUrl.Scheme
		request.URL.Host = targetUrl.Host

		trimmed := strings.TrimPrefix(request.URL.Path, path)
		if trimmed == "" {
			trimmed = "/"
		}
		request.URL.Path = trimmed

		request.Host = targetUrl.Host
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

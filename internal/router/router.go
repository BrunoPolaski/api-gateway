package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/BrunoPolaski/api-gateway/internal/entities"
	"github.com/BrunoPolaski/go-logger/logger"
	"github.com/BrunoPolaski/go-rest-err/rest_err"
)

func Init() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/", handleGateway)

	return r
}

func handleGateway(w http.ResponseWriter, r *http.Request) {
	for _, route := range routeTable {
		if (route.Method == "ANY" || route.Method == r.Method) && strings.HasPrefix(r.URL.Path, route.Path) {
			logger.Info(fmt.Sprintf("Forwarding %s %s%s", r.Method, route.Target, r.URL.Path))
			forwardRequest(w, r, route.Target)
			return
		}
	}
	http.NotFound(w, r)
}

var routeTable = []entities.Route{
	{
		Method: "ANY",
		Path:   "/sales",
		Target: "https://api-gateway.free.beeceptor.com",
	},
}

func forwardRequest(w http.ResponseWriter, r *http.Request, target string) {
	targetUrl, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
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

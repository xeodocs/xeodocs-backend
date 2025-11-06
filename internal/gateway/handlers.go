package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func AuthProxyHandler(cfg *config.Config) http.HandlerFunc {
	targetURL, _ := url.Parse(cfg.AuthServiceURL)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request to strip /v1/auth prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/v1")
		req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, "/v1")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For now, no JWT validation for auth endpoints
		proxy.ServeHTTP(w, r)
	}
}

package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// HandleEditorProxy proxies requests to Theia IDE
func HandleEditorProxy(w http.ResponseWriter, r *http.Request) {
	targetURL, err := url.Parse("http://localhost:9090")
	if err != nil {
		log.Printf("Failed to parse target URL: %v", err)
		http.Error(w, "Error parsing proxy URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = r.URL.Path
		req.Host = targetURL.Host

		// Handle WebSocket upgrade headers
		if r.Header.Get("Connection") == "Upgrade" && r.Header.Get("Upgrade") == "websocket" {
			log.Println("Proxying WebSocket upgrade request to Theia")
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
		}
	}

	proxy.ModifyResponse = func(res *http.Response) error {
		if loc, ok := res.Header["Location"]; ok && len(loc) > 0 {
			if strings.HasPrefix(loc[0], targetURL.String()) {
				res.Header.Set("Location", strings.Replace(loc[0], targetURL.String(), "/editor", 1))
			} else if strings.HasPrefix(loc[0], "/") {
				res.Header.Set("Location", "/editor"+loc[0])
			}
		}
		return nil
	}

	log.Printf("Proxying editor request for: %s", r.URL.Path)
	proxy.ServeHTTP(w, r)
}

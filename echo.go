package traefik_echoserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// Config the plugin configuration
type Config struct {
	Path                string `json:"path,omitempty"`
	AddHostnameToHeader bool   `json:"AddHostnameToHeader,omitempty"`
	Header              string `json:"header,omitempty"`
}

// CreateConfig creates the default plugin configuration
func CreateConfig() *Config {
	return &Config{
		Path:                "/cdn-cgi/info",
		AddHostnameToHeader: true,
		Header:              "Server",
	}
}

// echoServer
type echoServer struct {
	next                http.Handler
	name                string
	Path                string
	AddHostnameToHeader bool
	Header              string
}

// New created a new plugin
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &echoServer{
		next:                next,
		name:                name,
		Path:                config.Path,
		AddHostnameToHeader: config.AddHostnameToHeader,
		Header:              config.Header,
	}, nil
}

func getIP(r *http.Request) string {
	for _, i := range strings.Split(r.Header.Get("X-Forward-For"), ",") {
		if net.ParseIP(i) != nil {
			return i
		}
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}

func (r *echoServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.AddHostnameToHeader {
		if node, ok := os.LookupEnv("NODE_NAME"); ok {
			rw.Header().Add(r.Header, node)
		}
	}
	if req.URL.Path == r.Path {
		scheme := "http"
		tlsVersion := ""

		if req.TLS != nil {
			scheme = "https"
			switch req.TLS.Version {
			case tls.VersionTLS10:
				tlsVersion = "tls1.0"
			case tls.VersionTLS11:
				tlsVersion = "tls1.1"
			case tls.VersionTLS12:
				tlsVersion = "tls1.2"
			case tls.VersionTLS13:
				tlsVersion = "tls1.3"
			}
		}

		host := req.Host
		if strings.Contains(host, ":") {
			host, _, _ = net.SplitHostPort(host)
		}

		response := ""
		response += fmt.Sprintf("h=%s\n", host)
		response += fmt.Sprintf("ip=%s\n", getIP(req))
		response += fmt.Sprintf("ts=%d\n", time.Now().Local().UnixMilli())
		response += fmt.Sprintf("scheme=%s\n", scheme)
		response += fmt.Sprintf("http=%s\n", req.Proto)
		if tlsVersion != "" {
			response += fmt.Sprintf("tls=%s\n", tlsVersion)
		}
		if req.UserAgent() != "" {
			response += fmt.Sprintf("ua=%s\n", req.UserAgent())
		}
		if node, ok := os.LookupEnv("NODE_NAME"); ok {
			response += fmt.Sprintf("node=%s\n", node)
		}
		if pod, ok := os.LookupEnv("POD_NAME"); ok {
			response += fmt.Sprintf("pod=%s\n", pod)
		}
		response += fmt.Sprintf("arch=%s/%s\n", runtime.GOOS, runtime.GOARCH)

		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)

		if _, err := rw.Write([]byte(response)); err == nil {
			return
		}
	}

	if r.next != nil {
		r.next.ServeHTTP(rw, req)
	}
}

package traefik_echoserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestEchoServer(t *testing.T) {
	cfg := CreateConfig()

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "echo-server")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	reqUrl, err := url.JoinPath("http://localhost", cfg.Path)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	body, _ := io.ReadAll(resp.Body)

	t.Log(string(body))

	if recorder.Code != 200 {
		t.Errorf("code: %d", recorder.Code)
	}

	host := strings.Split(strings.Split(string(body), "\n")[0], "=")[1]
	if host != "localhost" {
		t.Errorf("host: %s", host)
	}
}

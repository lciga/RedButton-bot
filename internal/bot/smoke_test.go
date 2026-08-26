//go:build smoke

package bot

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"RedButton-bot/internal/service"
	telegram "github.com/go-telegram/bot"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (function httpClientFunc) Do(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestSmokeTelegramInitialization(t *testing.T) {
	client := httpClientFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/bottest-token/getMe" {
			t.Fatalf("unexpected Telegram path %q", request.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Smoke","username":"smoke_bot"}}`)),
			Request:    request,
		}, nil
	})

	application, err := newWithOptions(
		"test-token", &service.Services{}, nil, time.Second, time.Second, nil,
		time.Now().Add(-time.Hour), time.Now().Add(time.Hour),
		telegram.WithServerURL("http://telegram.invalid"), telegram.WithHTTPClient(time.Second, client),
	)
	if err != nil {
		t.Fatal(err)
	}
	if application.client == nil {
		t.Fatal("Telegram client was not initialized")
	}
}

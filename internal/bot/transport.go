package bot

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
)

type diagnosticTransport struct {
	base   *http.Transport
	logger *slog.Logger
}

type requestTrace struct {
	mutex                sync.Mutex
	dnsCompleted         bool
	connectAddress       string
	connectCompleted     bool
	connectError         error
	tlsHandshakesStarted int
	tlsHandshakesDone    int
	tlsError             error
	requestWritten       bool
	writeError           error
	firstResponseByte    bool
}

func (t *diagnosticTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	startedAt := time.Now()
	trace := &requestTrace{}
	request = request.Clone(httptrace.WithClientTrace(request.Context(), trace.clientTrace()))
	proxyURL, proxyError := t.base.Proxy(request)

	response, err := t.base.RoundTrip(request)
	attributes := append(
		[]any{
			"method", request.Method,
			"telegram_method", telegramMethod(request.URL),
			"duration", time.Since(startedAt),
		},
		trace.attributes()...,
	)
	attributes = append(attributes, proxyAttributes(proxyURL, proxyError)...)
	if err != nil {
		attributes = append(
			attributes,
			"error", err,
			"error_type", fmt.Sprintf("%T", err),
			"error_chain", errorChain(err),
		)
		t.logger.ErrorContext(request.Context(), "Telegram HTTP request failed", attributes...)
		return nil, err
	}

	t.logger.DebugContext(
		request.Context(),
		"Telegram HTTP request completed",
		append(attributes, "status_code", response.StatusCode)...,
	)
	return response, nil
}

func (t *requestTrace) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSDone: func(httptrace.DNSDoneInfo) {
			t.mutex.Lock()
			t.dnsCompleted = true
			t.mutex.Unlock()
		},
		ConnectStart: func(_, address string) {
			t.mutex.Lock()
			t.connectAddress = address
			t.mutex.Unlock()
		},
		ConnectDone: func(_, _ string, err error) {
			t.mutex.Lock()
			t.connectCompleted = true
			t.connectError = err
			t.mutex.Unlock()
		},
		TLSHandshakeStart: func() {
			t.mutex.Lock()
			t.tlsHandshakesStarted++
			t.mutex.Unlock()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			t.mutex.Lock()
			t.tlsHandshakesDone++
			t.tlsError = err
			t.mutex.Unlock()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			t.mutex.Lock()
			t.requestWritten = true
			t.writeError = info.Err
			t.mutex.Unlock()
		},
		GotFirstResponseByte: func() {
			t.mutex.Lock()
			t.firstResponseByte = true
			t.mutex.Unlock()
		},
	}
}

func (t *requestTrace) attributes() []any {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return []any{
		"dns_completed", t.dnsCompleted,
		"connect_address", t.connectAddress,
		"connect_completed", t.connectCompleted,
		"connect_error", t.connectError,
		"tls_handshakes_started", t.tlsHandshakesStarted,
		"tls_handshakes_done", t.tlsHandshakesDone,
		"tls_error", t.tlsError,
		"request_written", t.requestWritten,
		"write_error", t.writeError,
		"first_response_byte", t.firstResponseByte,
	}
}

func logProxyConfiguration(logger *slog.Logger, transport *http.Transport) {
	request := &http.Request{URL: &url.URL{Scheme: "https", Host: "api.telegram.org"}}
	proxyURL, err := transport.Proxy(request)
	logger.Info("Telegram HTTP transport configured", proxyAttributes(proxyURL, err)...)
}

func proxyAttributes(proxyURL *url.URL, err error) []any {
	attributes := []any{"proxy_enabled", proxyURL != nil, "proxy_error", err}
	if proxyURL == nil {
		return attributes
	}

	return append(
		attributes,
		"proxy_scheme", proxyURL.Scheme,
		"proxy_host", proxyURL.Hostname(),
		"proxy_port", proxyURL.Port(),
		"proxy_auth_configured", proxyURL.User != nil,
	)
}

func telegramMethod(requestURL *url.URL) string {
	path := strings.TrimSuffix(requestURL.Path, "/")
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

func errorChain(err error) string {
	chain := make([]string, 0, 4)
	for err != nil {
		chain = append(chain, fmt.Sprintf("%T", err))
		err = errors.Unwrap(err)
	}
	return strings.Join(chain, " -> ")
}

var _ http.RoundTripper = (*diagnosticTransport)(nil)

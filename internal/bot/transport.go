package bot

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"

	applicationlogger "RedButton-bot/internal/logger"
)

type diagnosticTransport struct {
	base   *http.Transport
	logger *slog.Logger
}

type requestTrace struct {
	mutex                sync.Mutex
	connectAddress       string
	connectCompleted     bool
	connectError         error
	tlsHandshakesStarted int
	tlsHandshakesDone    int
	tlsProtocols         []string
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
			"operation", telegramMethod(request.URL),
			"duration", time.Since(startedAt),
		},
		trace.attributes()...,
	)
	attributes = append(attributes, proxyAttributes(proxyURL, proxyError)...)
	if err != nil {
		attributes = append(
			attributes,
			"error", err,
		)
		t.logger.ErrorContext(request.Context(), "Telegram HTTP request failed", attributes...)
		return nil, applicationlogger.MarkLogged(err)
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
		TLSHandshakeDone: func(state tls.ConnectionState, _ error) {
			t.mutex.Lock()
			t.tlsHandshakesDone++
			t.tlsProtocols = append(t.tlsProtocols, state.NegotiatedProtocol)
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
		"connect_address", t.connectAddress,
		"connected", t.connectCompleted && t.connectError == nil,
		"tls_handshakes", fmt.Sprintf("%d/%d", t.tlsHandshakesDone, t.tlsHandshakesStarted),
		"tls_protocols", strings.Join(t.tlsProtocols, ","),
		"request_written", t.requestWritten && t.writeError == nil,
		"response_started", t.firstResponseByte,
	}
}

func logProxyConfiguration(logger *slog.Logger, transport *http.Transport) {
	request := &http.Request{URL: &url.URL{Scheme: "https", Host: "api.telegram.org"}}
	proxyURL, err := transport.Proxy(request)
	logger.Debug("Telegram transport configured", proxyAttributes(proxyURL, err)...)
}

func proxyAttributes(proxyURL *url.URL, err error) []any {
	attributes := make([]any, 0, 4)
	if err != nil {
		attributes = append(attributes, "proxy_error", err)
	}
	if proxyURL == nil {
		return append(attributes, "proxy", "direct")
	}

	return append(
		attributes,
		"proxy", fmt.Sprintf("%s://%s", proxyURL.Scheme, proxyURL.Host),
	)
}

func telegramMethod(requestURL *url.URL) string {
	path := strings.TrimSuffix(requestURL.Path, "/")
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

var _ http.RoundTripper = (*diagnosticTransport)(nil)

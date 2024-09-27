package proxy

import (
	"io"
	"net/http"
	"strings"
)

type HeaderHandler struct {

	//TODO: handle keep alive
	//# https://stackoverflow.com/questions/40824363/http-proxy-server-keep-alive-connection-support
	hopHeaders []string
}

func NewHeaderHandler() *HeaderHandler {
	hh := &HeaderHandler{
		hopHeaders: []string{
			"Connection",
			"Keep-Alive",
			"Proxy-Authenticate",
			"Proxy-Authorization",
			"Te", // canonicalized version of "TE"
			"Trailers",
			"Transfer-Encoding",
			"Upgrade",
		},
	}

	return hh
}

func (hh HeaderHandler) delHopHeaders(header http.Header) {
	for _, h := range hh.hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (hh HeaderHandler) PreRequest(originalRequest *http.Request) {
	originalRequest.RequestURI = ""

	hh.delHopHeaders(originalRequest.Header)
}

func (hh HeaderHandler) PostRequest(proxyResp *http.Response, originalWriter http.ResponseWriter) {
	defer proxyResp.Body.Close()
	hh.delHopHeaders(proxyResp.Header)

	copyHeader(originalWriter.Header(), proxyResp.Header)
	// originalWriter.WriteHeader(proxyResp.StatusCode)
	io.Copy(originalWriter, proxyResp.Body)
}

package proxy

import (
	"io"
	"net"
	"net/http"
)

type RequestHandler struct {
	client *http.Client
}

func (r RequestHandler) Do(w http.ResponseWriter, originalRequest *http.Request) (resp *http.Response, httpErr error) {
	if clientIP, _, err := net.SplitHostPort(originalRequest.RemoteAddr); err == nil {
		appendHostToXForwardHeader(originalRequest.Header, clientIP)
	}

	// //TODO: Add timeout
	resp, httpErr = http.DefaultTransport.RoundTrip(originalRequest)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers and status code from the target server response
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy the response body to the client
	io.Copy(w, resp.Body)

	return resp, httpErr
}

func NewRequestHandler() *RequestHandler {
	r := &RequestHandler{
		client: &http.Client{},
	}

	return r
}

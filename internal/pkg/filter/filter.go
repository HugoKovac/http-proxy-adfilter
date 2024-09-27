package filter

import (
	"errors"
	"net/http"
)

func Filter(originalRequest *http.Request) error {
	if originalRequest.URL.Scheme != "http" { // TODO: inplement ws and https
		return errors.New("Scheme not supported")
	}
	//TODO: Filter domain
	// check if client is registered and has ip and domain blocked
	return nil
}

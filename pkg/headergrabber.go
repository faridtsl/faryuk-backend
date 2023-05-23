package pkg

import (
	"crypto/tls"
	"net/http"
)

// HeaderGrabber : struct for grabbing headers
type HeaderGrabber struct{}

// NewHeaderGrabber : returns new HeaderGrabber
func NewHeaderGrabber() *HeaderGrabber {
	return &HeaderGrabber{}
}

// Run : gets HTTP headers
func (h HeaderGrabber) Run(url string) (map[string][]string, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Head(url)

	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}
	return headers, nil
}

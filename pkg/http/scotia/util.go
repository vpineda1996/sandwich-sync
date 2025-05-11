package scotia

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func debugRoundTripper() http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		d, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(d))
		res, err := http.DefaultTransport.RoundTrip(r)
		if err == nil {
			d, _ := httputil.DumpResponse(res, true)
			fmt.Println(string(d))
		}
		return res, err
	})
}

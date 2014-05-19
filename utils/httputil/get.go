// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package httputil

import (
	"fmt"
	"net/http"
)

func Get(URL string, cred Credentials) (*http.Response, error) {
	// Prepare the HTTP request.
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	if cred != nil {
		req.SetBasicAuth(cred.Username(), cred.Password())
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

type transport struct {
	rt http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("User-Agent", "curl/7.21.4 (universal-apple-darwin11.0) libcurl/7.21.4 OpenSSL/0.9.8y zlib/1.2.5")
	req.Header.Set("Accept", "*/*")
	fmt.Println("-------")
	fmt.Println("REQUEST", req)
	resp, err = t.rt.RoundTrip(req)
	fmt.Println("-------")
	fmt.Println("RESPONSE")
	fmt.Println(resp)
	fmt.Println("ERROR")
	fmt.Println(err)
	fmt.Println("-------")
	return
}

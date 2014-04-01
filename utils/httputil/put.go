// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package httputil

import (
	"io"
	"net/http"
	"os"
)

func Put(body io.Reader, URL string, cred Credentials) (*http.Response, error) {
	// Prepare the HTTP request.
	req, err := http.NewRequest("PUT", URL, body)
	if err != nil {
		return nil, err
	}
	if cred != nil {
		req.SetBasicAuth(cred.Username(), cred.Password())
	}

	// Try to set Content-Length in some more special cases.
	switch v := body.(type) {
	case *os.File:
		info, err := v.Stat()
		if err != nil {
			return nil, err
		}

		req.ContentLength = info.Size()
	}

	// Send the request.
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return resp, nil
}

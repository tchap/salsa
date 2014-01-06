// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package httputil

import (
	"io"
	"net/http"
)

type Credentials interface {
	Username() string
	Password() string
}

func Put(body io.Reader, URL string, cred Credentials) (*http.Response, error) {
	// Prepare the HTTP request.
	req, err := http.NewRequest("PUT", URL, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cred.Username(), cred.Password())

	// Send the request.
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return resp, nil
}

// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package httputil

import "net/http"

func Get(URL string, cred Credentials) (*http.Response, error) {
	// Prepare the HTTP request.
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	if cred != nil {
		req.SetBasicAuth(cred.Username(), cred.Password())
	}

	// Send the request.
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package httputil

type Credentials interface {
	Username() string
	Password() string
}

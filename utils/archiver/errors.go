// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package archiver

import "errors"

var (
	ErrUnknownArchiverType = errors.New("Unknown archiver type")
	ErrNoArtifacts         = errors.New("No artifacts found")
)

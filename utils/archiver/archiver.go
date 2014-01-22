// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package archiver

import "os"

type Options interface {
	Verbose() bool
	Dry() bool
}

type Archiver interface {
	Archive(srcDir string) (archive *os.File, err error)
}

type ArchiverType string

const (
	TgzArchiverType ArchiverType = "tar.gz"
)

func New(typ ArchiverType, opts Options) (Archiver, error) {
	switch typ {
	case TgzArchiverType:
		return newTgzArchiver(opts), nil
	}

	return nil, ErrUnknownArchiverType
}

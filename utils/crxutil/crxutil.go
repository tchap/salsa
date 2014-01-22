// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package crxutil

import (
	"encoding/binary"
	"errors"
	"io"
)

const MagicNumber = "Cr24"

var ErrNotCrx = errors.New("not a crx file")

type CrxFile struct {
	MagicNumber string
	Version     uint32
	PublicKey   []byte
	Signature   []byte
	ZipFile     io.ReadCloser
}

func NewCrxFile(rc io.ReadCloser) (*CrxFile, error) {
	crx := new(CrxFile)
	var (
		magicNumber  []byte = make([]byte, 4)
		publicKeyLen uint32
		signatureLen uint32
	)

	// Read magic number.
	for i := 0; i != len(magicNumber); {
		n, err := rc.Read(magicNumber[i:])
		if err != nil {
			return nil, err
		}
		i += n
	}
	crx.MagicNumber = string(magicNumber)
	if crx.MagicNumber != MagicNumber {
		return nil, ErrNotCrx
	}

	// Read version.
	if err := binary.Read(rc, binary.LittleEndian, &crx.Version); err != nil {
		return nil, err
	}

	// Read public key length.
	if err := binary.Read(rc, binary.LittleEndian, &publicKeyLen); err != nil {
		return nil, err
	}

	// Read signature length.
	if err := binary.Read(rc, binary.LittleEndian, &signatureLen); err != nil {
		return nil, err
	}

	// Read the public key.
	crx.PublicKey = make([]byte, publicKeyLen)
	for i := uint32(0); i != publicKeyLen; {
		n, err := rc.Read(crx.PublicKey[i:])
		if err != nil {
			return nil, err
		}
		i += uint32(n)
	}

	// Read the signature.
	crx.Signature = make([]byte, signatureLen)
	for i := uint32(0); i != signatureLen; {
		n, err := rc.Read(crx.Signature[i:])
		if err != nil {
			return nil, err
		}
		i += uint32(n)
	}

	// What is left is the zip file.
	crx.ZipFile = rc
	return crx, nil
}

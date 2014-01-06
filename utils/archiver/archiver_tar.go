// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package archiver

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type tarArchiver struct {
	opts Options
}

func newTarArchiver(opts Options) *tarArchiver {
	return &tarArchiver{opts}
}

func (archiver *tarArchiver) Archive(srcDir string) (archive *os.File, err error) {
	// Make sure the artifacts source directory exists and is not empty.
	dir, err := os.Open(srcDir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	info, err := dir.Readdir(1)
	if err != nil {
		return nil, err
	}

	if len(info) == 0 {
		return nil, ErrNoArtifacts
	}

	// Pack the artifacts directory.
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	ar, err := ioutil.TempFile(wd, "artifacts_archive_")
	if err != nil {
		return nil, err
	}

	if archiver.opts.Verbose() {
		fmt.Println("Packing artifacts")
	}

	aw := tar.NewWriter(ar)

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		// Stop on error.
		if err != nil {
			return err
		}

		// Skip root.
		if path == srcDir {
			return nil
		}

		if archiver.opts.Verbose() {
			fmt.Println("   ", path)
		}

		// Open the artifacts file.
		relative := path[len(srcDir):]
		file, err := os.Open(filepath.Join(srcDir, relative))
		if err != nil {
			return err
		}
		defer file.Close()

		// Prepare tar header.
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		if archiver.opts.Dry() {
			header.Size = 0
		}

		// Write tar header.
		if err := aw.WriteHeader(header); err != nil {
			return err
		}

		// Copy the file into the archive.
		if !archiver.opts.Dry() {
			_, err = io.Copy(aw, file)
		}
		return err
	})
	if err != nil {
		aw.Close()
		ar.Close()
		return nil, err
	}

	if archiver.opts.Verbose() {
		fmt.Println("Archive created")
	}

	// Make sure we close the archive writer properly.
	if err := aw.Close(); err != nil {
		ar.Close()
		return nil, err
	}

	// Rewind to the beginning of the archive, otherwise the following reads
	// will return no data at all.
	if _, err := ar.Seek(0, os.SEEK_SET); err != nil {
		ar.Close()
		return nil, err
	}

	// Return the archive file, open and set to offset 0.
	return ar, nil
}

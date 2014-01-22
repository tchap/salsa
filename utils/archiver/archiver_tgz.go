// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package archiver

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"
)

type tgzArchiver struct {
	opts Options
}

func newTgzArchiver(opts Options) *tgzArchiver {
	return &tgzArchiver{opts}
}

func (archiver *tgzArchiver) Archive(srcDir string) (archive *os.File, err error) {
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

	gzipWriter := gzip.NewWriter(ar)
	tarWriter := tar.NewWriter(gzipWriter)

	if archiver.opts.Verbose() {
		fmt.Println("Packing artifacts")
	}

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		// Stop on error.
		if err != nil {
			return err
		}

		relative := path[len(srcDir):]

		// Skip root.
		if len(relative) == 0 {
			return nil
		}

		// Drop the leading slash.
		if r, _ := utf8.DecodeRuneInString(relative); r == os.PathSeparator {
			relative = relative[1:]
		}

		// Prepare tar header.
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Append a trailing slash if this is a directory.
		if info.IsDir() {
			relative = fmt.Sprintf("%v%c", filepath.Clean(relative), os.PathSeparator)
		}

		if archiver.opts.Verbose() {
			fmt.Println("   ", relative)
		}

		// Tar always uses '/' as the separator.
		header.Name = filepath.ToSlash(relative)

		if archiver.opts.Dry() {
			header.Size = 0
		}

		// Write tar header.
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Do not write directories.
		if info.IsDir() {
			return nil
		}

		// Open the artifacts file.
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file into the archive.
		if !archiver.opts.Dry() {
			_, err = io.Copy(tarWriter, file)
		}
		return err
	})
	if err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		ar.Close()
		os.Remove(ar.Name())
		return nil, err
	}

	if archiver.opts.Verbose() {
		fmt.Println("Archive created")
	}

	// Make sure we close tar writer properly.
	if err := tarWriter.Close(); err != nil {
		gzipWriter.Close()
		ar.Close()
		os.Remove(ar.Name())
		return nil, err
	}

	// Make sure we close gzip writer properly.
	if err := gzipWriter.Close(); err != nil {
		ar.Close()
		os.Remove(ar.Name())
		return nil, err
	}

	// Rewind to the beginning of the archive, otherwise the following reads
	// will return no data at all.
	if _, err := ar.Seek(0, os.SEEK_SET); err != nil {
		ar.Close()
		os.Remove(ar.Name())
		return nil, err
	}

	// Return the archive file, open and set to offset 0.
	return ar, nil
}

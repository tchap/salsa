// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	// Stdlib
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	// Salsa
	"github.com/tchap/salsa/utils/httputil"

	// Others
	"github.com/tchap/gocli"
)

const crxURLTemplate = "https://clients2.google.com/service/update2/crx?response=redirect&x=id%3D~~~~%26uc"

// Subcommand initialisation and registration.
func init() {
	chromeExtension := &gocli.Command{
		UsageLine: `
  chrome_extension SUBCMD`,
		Short: "manipulate Chrome Web Store extensions",
	}

	getCrx := &gocli.Command{
		UsageLine: `
  get_crx EXTENSION_ID FILENAME`,
		Short: "download Chrome extensions from Chrome Web Store",
		Long: `
  Download the extensions identified by EXTENSION_ID and save it in FILENAME.
		`,
		Action: runGetCrx,
	}
	chromeExtension.MustRegisterSubcommand(getCrx)

	getApp().MustRegisterSubcommand(chromeExtension)
}

// Subcommand handler.
func runGetCrx(cmd *gocli.Command, args []string) {
	if len(args) != 2 {
		cmd.Usage()
		os.Exit(2)
	}

	var (
		id       = args[0]
		filename = args[1]
	)

	URL := strings.Replace(crxURLTemplate, "~~~~", id, 1)

	if config.Verbose() {
		fmt.Println("GET", URL)
	}
	if config.Dry() {
		return
	}

	// Download CRX.
	resp, err := httputil.Get(URL, nil)
	if err != nil {
		log.Fatalf("Error: failed to download crx: %v\n", err)
	}
	if resp.StatusCode >= 300 {
		log.Fatalf("Error: failed to download crx: %v\n", resp.Status)
	}
	defer resp.Body.Close()

	// Write it to the file.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	defer file.Close()

	n, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Printf("Wrote %v bytes to %v\n", n, filename)
}

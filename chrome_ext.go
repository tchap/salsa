// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	// Stdlib
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	// Salsa
	"github.com/tchap/salsa/utils/flagutil"
	"github.com/tchap/salsa/utils/httputil"

	// Others
	"github.com/tchap/gocli"
)

const crxURLTemplate = "https://clients2.google.com/service/update2/crx?response=redirect&x=id%3D~~~~%26uc"

// Subcommand initialisation and registration.
func init() {
	chromeExt := &gocli.Command{
		UsageLine: `
  chrome_ext SUBCMD`,
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
	chromeExt.MustRegisterSubcommand(getCrx)

	genPackageJson := &gocli.Command{
		UsageLine: `
  gen_package_json [-dep NAME:VERSION ...] MANIFEST_FILE`,
		Short: "generate package.json from manifest.json",
		Long: `
  Generate package.json in the current working directory from manifest.json
  living at MANIFEST_FILE.

  Project name and version from manifest.json is reused, dependencies can be
  added to package.json using -dep option, which can be used multiple times.
  NAME is used as the dependency map key, version as the value
		`,
		Action: runGenPackageJson,
	}
	genPackageJson.Flags.Var(packageJsonDeps, "dep", "add a dependency into package.json")
	chromeExt.MustRegisterSubcommand(genPackageJson)

	getApp().MustRegisterSubcommand(chromeExt)
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

	fmt.Printf("Wrote %v bytes\n", n)
}

var packageJsonDeps = flagutil.NewMapValue()

// Subcommand handler.
func runGenPackageJson(cmd *gocli.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	var manifestFilename = args[0]

	content, err := ioutil.ReadFile(manifestFilename)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	var packageJson struct {
		Name         string            `json:"name"`
		Version      string            `json:"version"`
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(content, &packageJson); err != nil {
		log.Fatalf("Error: failed to unmarshal manifest.json: %v", err)
	}

	packageJson.Dependencies = packageJsonDeps.M

	content, err = json.MarshalIndent(packageJson, "", "  ")
	if err != nil {
		log.Fatalf("Error: failed to marshal package.json: %v", err)
	}

	file, err := os.OpenFile("package.json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(content); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("package.json created")
}
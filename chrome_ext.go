// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	// Stdlib
	"encoding/binary"
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

	getName := &gocli.Command{
		UsageLine: `
  get_name MANIFEST_FILE`,
		Short:  "get name string from a manifest file",
		Action: getNameFromManifest,
	}
	chromeExt.MustRegisterSubcommand(getName)

	getVersion := &gocli.Command{
		UsageLine: `
  get_version MANIFEST_FILE`,
		Short:  "get version string from a manifest file",
		Action: getVersionFromManifest,
	}
	chromeExt.MustRegisterSubcommand(getVersion)

	getCrx := &gocli.Command{
		UsageLine: `
  get_crx [-zip] [-url=URL] [EXTENSION_ID] FILENAME`,
		Short: "download Chrome extensions from Chrome Web Store",
		Long: `
  Download the extensions identified by EXTENSION_ID and save it in FILENAME.
  The -url flag can be used to download the package from arbitrary location.
		`,
		Action: runGetCrx,
	}
	getCrx.Flags.BoolVar(&convertCrxToZip, "zip", convertCrxToZip, "convert crx to zip")
	getCrx.Flags.StringVar(&crxURL, "url", crxURL, "CRX package URL")
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

var (
	convertCrxToZip bool
	crxURL          string
)

// Subcommand handler.
func runGetCrx(cmd *gocli.Command, args []string) {
	var (
		id       string
		filename string
	)
	if len(args) == 2 {
		id = args[0]
		filename = args[1]
	} else if len(args) == 1 && crxURL != "" {
		filename = args[0]
	} else {
		cmd.Usage()
		os.Exit(2)
	}

	var packageURL string
	if crxURL != "" {
		packageURL = crxURL
	} else {
		packageURL = strings.Replace(crxURLTemplate, "~~~~", id, 1)
	}

	if config.Verbose() {
		fmt.Println("GET", packageURL)
	}
	if config.Dry() {
		return
	}

	// Download CRX.
	resp, err := httputil.Get(packageURL, nil)
	if err != nil {
		log.Fatalf("Error: failed to download crx: %v\n", err)
	}
	if resp.StatusCode >= 300 {
		log.Fatalf("Error: failed to download crx: %v\n", resp.Status)
	}
	defer resp.Body.Close()

	// Convert CRX to ZIP if requested.
	if convertCrxToZip {
		var (
			publicKeyLen uint32
			signatureLen uint32
		)

		// Drop magic number.
		if err := binary.Read(resp.Body, binary.LittleEndian, &publicKeyLen); err != nil {
			log.Fatal(err)
		}
		// Drop version.
		if err := binary.Read(resp.Body, binary.LittleEndian, &publicKeyLen); err != nil {
			log.Fatal(err)
		}
		// Read public key length.
		if err := binary.Read(resp.Body, binary.LittleEndian, &publicKeyLen); err != nil {
			log.Fatal(err)
		}
		// Read signature length.
		if err := binary.Read(resp.Body, binary.LittleEndian, &signatureLen); err != nil {
			log.Fatal(err)
		}

		var p []byte
		if publicKeyLen > signatureLen {
			p = make([]byte, publicKeyLen)
		} else {
			p = make([]byte, signatureLen)
		}

		// Drop the public key.
		b := p[:publicKeyLen]
		for i := uint32(0); i != publicKeyLen; {
			n, err := resp.Body.Read(b[i:])
			if err != nil {
				log.Fatal(err)
			}
			i += uint32(n)
		}
		// Drop the signature.
		b = p[:signatureLen]
		for i := uint32(0); i != signatureLen; {
			n, err := resp.Body.Read(b[i:])
			if err != nil {
				log.Fatal(err)
			}
			i += uint32(n)
		}
	}

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

	packageJson.Name = strings.ToLower(packageJson.Name)
	packageJson.Name = strings.Replace(packageJson.Name, " ", "-", -1)
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

func getNameFromManifest(cmd *gocli.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	manifest, err := loadManifest(args[0])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Print(manifest.Name)
}

func getVersionFromManifest(cmd *gocli.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	manifest, err := loadManifest(args[0])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Print(manifest.Version)
}

type manifest struct {
	Name    string
	Version string
}

func loadManifest(filename string) (*manifest, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var m manifest
	if err := json.Unmarshal(content, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

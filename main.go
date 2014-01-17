// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	// Stdlib
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	// Others
	"github.com/tchap/gocli"
)

const (
	SalsaRCFile = ".salsarc"
	PackageFile = "package.json"

	VersionPattern = "^[0-9]+[.][0-9]+[.][0-9]+$"
)

// Config is encapsulating configuration as collected from various sources,
// namely package.json, .salsarc and command line flags.
type Config struct {
	Package struct {
		Name    string
		Version string
	}
	RC struct {
		StoreURL string `json:"storeURL"`
		Secrets  map[string]string
		Username string
		Password string
	}
	Flags struct {
		Verbose  bool
		Dry      bool
		Username string
		Password string
	}
}

func (config *Config) Verbose() bool {
	return config.Flags.Verbose
}

func (config *Config) Dry() bool {
	return config.Flags.Dry
}

func (config *Config) Username() string {
	return config.RC.Username
}

func (config *Config) Password() string {
	return config.RC.Password
}

// Global config instance that is used to collect command line flags.
// The default config can be set by setting fields in this object.
var config = new(Config)

func bootstrap() {
	// Load package.json first.
	if config.Verbose() {
		fmt.Printf("Reading %v ...\n", PackageFile)
	}

	content, err := ioutil.ReadFile(PackageFile)
	if err != nil {
		log.Fatalf("Error: failed to read %v: %v", PackageFile, err)
	}
	if err := json.Unmarshal(content, &config.Package); err != nil {
		log.Fatalf("Error: failed to unmarshal %v: %v", PackageFile, err)
	}

	// Update the config in cascade, $HOME/.salsarc -> $PWD/.salsarc
	user, err := user.Current()
	if err != nil {
		log.Fatalf("Error: failed to get the current user: %v", err)
	}

	var userConfig string
	if config := os.Getenv("SALSA_RC"); config != "" {
		userConfig = config
	} else {
		userConfig = filepath.Join(user.HomeDir, SalsaRCFile)
	}
	checkPermissions := os.Getenv("SALSA_SKIP_PERMISSIONS_CHECK") == ""
	rcFiles := []string{
		userConfig,
		SalsaRCFile,
	}
	for _, rcPath := range rcFiles {
		if config.Verbose() {
			fmt.Printf("Reading %v ...\n", rcPath)
		}

		if checkPermissions && rcPath != SalsaRCFile {
			info, err := os.Stat(rcPath)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Fatalf("Error: failed to stat %v: %v", rcPath, err)
				}
			}
			if perm := info.Mode() & os.ModePerm & 0007; perm != 0 {
				log.Fatalf("Error: %v violates the permissions constraints", rcPath)
			}
		}

		content, err := ioutil.ReadFile(rcPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("Error: failed to read %v: %v", rcPath, err)
			}
			continue
		}

		if err := json.Unmarshal(content, &config.RC); err != nil {
			log.Fatalf("Error: failed to unmarshal %v: %v", rcPath, err)
		}
	}

	// Verify the config.
	switch {
	case config.Package.Name == "":
		log.Fatalln("Error: empty package name")
	case config.RC.Secrets[config.Package.Name] == "":
		log.Fatalf("Error: secret not found for project %v", config.Package.Name)
	}
	match, err := regexp.Match(VersionPattern, []byte(config.Package.Version))
	if err != nil {
		panic(err)
	}
	if !match {
		log.Fatalln("Error: version format mismatch")
	}

	// Set the credentials as expected, that is Flags overwrite all.
	if config.Flags.Username != "" {
		config.RC.Username = config.Flags.Username
	}
	if config.Flags.Password != "" {
		config.RC.Password = config.Flags.Password
	}
}

// gocli App for parsing of the command line.
var app *gocli.App

func getApp() *gocli.App {
	// Return the app if already initialised.
	if app != nil {
		return app
	}

	// Otherwise initialise the app and return the new instance.
	app = gocli.NewApp("salsa")
	app.UsageLine = `
  salsa [-h] [-v] [-dry] [-username USER -password PASSWD] SUBCMD`
	app.Short = "a project build artifacts manager"
	app.Version = "0.0.1"
	app.Long = `
  Salsa can be used to upload or download project build artifacts to or from
  a remote HTTP server using HTTP GET/PUT respectively.

  Salsa can also use Basic auth to authenticate HTTP requests. If you, however,
  do not want to use your username and password in CLI, create .salsarc in your
  home directory. It shall be a json file containing "username" and "password".
  If this file is found, salsa will read the credentials from there.`
	app.Flags.BoolVar(&config.Flags.Verbose, "v", config.Flags.Verbose,
		"print verbose output")
	app.Flags.BoolVar(&config.Flags.Dry, "dry", config.Flags.Dry,
		"just print what would be executed")
	app.Flags.StringVar(&config.RC.Username, "username", config.RC.Username,
		"Basic auth username")
	app.Flags.StringVar(&config.RC.Password, "password", "",
		"Basic auth password")

	return app
}

func main() {
	log.SetFlags(0)
	getApp().Run(os.Args[1:])
}

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
	ConfigFilename = ".salsarc"
	PackageFile    = "package.json"

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
	// Part I: Load package.json first.
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

	// Part II: Update config in cascade from $HOME/.salsarc, then $PWD/.salsarc
	user, err := user.Current()
	if err != nil {
		log.Fatalf("Error: failed to get the current user: %v", err)
	}

	userConfig := os.Getenv("SALSA_USER_CONFIG")
	if userConfig == "" {
		userConfig = filepath.Join(user.HomeDir, ConfigFilename)
	}

	// Print warning if the user-specific config file is accessible by other
	// users. Its mode should be set to 0600 since it can containt credentials.
	if info, err := os.Stat(userConfig); err == nil {
		if perm := info.Mode() & os.ModePerm & 0077; perm != 0 {
			fmt.Println("WARNING: %v is accessible by other users")
		}
	} else {
		if !os.IsNotExist(err) {
			log.Fatalf("Error: failed to stat %v: %v", userConfig, err)
		}
	}

	// Read and unmarshal the config files in cascade.
	// $PWD/.salsarc overwrites $HOME/.salsarc
	for _, configFile := range []string{userConfig, ConfigFilename} {
		if config.Verbose() {
			fmt.Printf("Reading %v ...\n", configFile)
		}

		content, err := ioutil.ReadFile(configFile)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("Error: failed to read %v: %v", configFile, err)
			}
			continue
		}

		if err := json.Unmarshal(content, &config.RC); err != nil {
			log.Fatalf("Error: failed to unmarshal %v: %v", configFile, err)
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
  Salsa is a project build artifacts manager that can publish or fetch build
  artifacts. A remote HTTP server acts as the artifacts store and salsa uses
  HTTP PUT and GET requests to publish and fetch the artifacts respectively.
  See the subcommands for more details.

  Salsa can be set up to use Basic authentication to authenticate HTTP requests.
  If you, however, do not want to specify the credentials on the command line,
  $HOME/.salsarc can be used to set them for you.

ENVIRONMENTAL VARIABLES:
  SALSA_USER_CONFIG - overwrites the default location for the user-specific
                      configuration file, which is $HOME/.salsarc`
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

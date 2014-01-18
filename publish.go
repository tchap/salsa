// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	// Stdlib
	"fmt"
	"log"
	"os"
	"strings"

	// Salsa
	"github.com/tchap/salsa/utils/archiver"
	"github.com/tchap/salsa/utils/httputil"

	// Others
	"github.com/tchap/gocli"
)

// Subcommand flags.
var (
	publishTag      string
	publishArchiver string = "tar"
)

// Subcommand initialisation and registration.
func init() {
	publish := &gocli.Command{
		UsageLine: `
  publish [-tag TAG] [-archiver {tar|zip}] ARTIFACTS_DIR`,
		Short: "publish build artifacts",
		Long: `
  publish uses ARTIFACTS_DIR as the root directory for the archive that it
  creates and uploads to the server.

  publish goes through the following steps:
    1. read package.json in the current working directory (mandatory),
    2. read .salsarc in the current working directory (optional),
    3. read the user-specific salsa config file (mandatory),
    4. create the archive from ARTIFACTS_DIR using the selected archiver,
    5. PUT the archive to $storeURL/$project-$secret/$branch/$archive where
       archive=$project-$tag-$branch-$version.$archiver

  All the configuration files are JSON files containing relevant keys:
    * package.json is the NPM package.json, salsa uses "name" and "version"
    * .salsarc can contain a project-specific artifacts store as "storeURL"
    * the user-specific .salsarc can contain "storeURL" as well as the HTTP
      Basic authentication credentials as "username" and "password".
      Project URL secrets are also store there under "secrets.$project"

ENVIRONMENTAL VARIABLES:
  BRANCH       - if set, $BRANCH is used in the archive filename as $branch
  BUILD_NUMBER - if set, $version is set to $version.$BUILD_NUMBER
		`,
		Action: runPublish,
	}

	publish.Flags.StringVar(&publishTag, "tag", publishTag,
		"tag to use in the archive file name")
	publish.Flags.StringVar(&publishArchiver, "archiver", publishArchiver,
		"archiver to use for packing the artifacts")

	getApp().MustRegisterSubcommand(publish)
}

// Subcommand handler.
func runPublish(cmd *gocli.Command, args []string) {
	// Update publisher config depending on the command line form that was used.
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	// Load the configuration.
	bootstrap()

	// Read the environment.
	branch := os.Getenv("BRANCH")
	if branch == "" {
		branch = "unknown"
	} else {
		branch = strings.Replace(branch, "/", "", -1)
	}
	if buildNum := os.Getenv("BUILD_NUMBER"); buildNum != "" {
		config.Package.Version += "." + buildNum
	}

	// Pack the matching artifacts into an archive.
	archiver, err := archiver.New(archiver.ArchiverType(publishArchiver), config)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	archive, err := archiver.Archive(args[0])
	if err != nil {
		log.Fatalf("Error: failed to create the artifacts archive: %v", err)
	}

	var exitError error
	defer func() {
		if err := os.Remove(archive.Name()); err != nil {
			log.Printf("Warning: failed to remove temporary file %v: %v",
				archive.Name(), err)
		}
		if exitError != nil {
			log.Fatal(exitError)
		}
	}()

	// Upload the archive.
	if publishTag != "" {
		publishTag = "-" + publishTag
	}
	filename := fmt.Sprintf(
		"%v%v-%v-%v.%v",
		config.Package.Name,
		publishTag,
		branch,
		config.Package.Version,
		publishArchiver)
	URL := fmt.Sprintf(
		"%v/%v-%v/%v/%v",
		config.RC.StoreURL,
		config.Package.Name,
		config.RC.Secrets[config.Package.Name],
		branch,
		filename)

	if config.Verbose() {
		fmt.Printf("PUT %v\n", URL)
	}
	if config.Dry() {
		fmt.Printf("Archive uploaded to\n\n  %v\n\n", URL)
		return
	}

	resp, err := httputil.Put(archive, URL, config)
	if err != nil {
		exitError = fmt.Errorf("Error: failed to upload the archive: %v", err)
		return
	}
	if resp.StatusCode >= 300 {
		exitError = fmt.Errorf("Error: failed to upload the archive: %v", resp.Status)
		return
	}

	fmt.Printf("Archive uploaded to\n\n  %v\n\n", URL)
}

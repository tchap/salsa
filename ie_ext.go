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
	"regexp"
	"strings"
	"text/template"

	// Others
	"github.com/dmotylev/nutrition"
	"github.com/tchap/gocli"
)

// Subcommand initialisation and registration.
func init() {
	ieExt := &gocli.Command{
		UsageLine: `
  ie_ext SUBCMD`,
		Short: "IE extensions manipulation",
	}

	genBhoversionRc := &gocli.Command{
		UsageLine: `
  gen_bhoversion_rc [-manifest MANIFEST_FILE] [FILE]`,
		Short: "generate bhoversion.rc",
		Long: `
  gen_brhoversion_rc generates bhoversion.rc in the current working directory
  unless FILE is specified. It can optionally read a Chrome extension
  manifest.json to get the extension version.

  This subcommand uses environmental variables to fill the bhoversion.rc template.
  The required environmental variables are:
    BHRC_COMPANYNAME
    BHRC_FILEDESCRIPTION
    BHRC_VERSION - must be a.b.c.d, $BUILD_NUMBER is used for $d if present;
                   does not have to be set if manifest is being used
    BHRC_LEGALCOPYRIGHT
    BHRC_PRODUCTNAME
		`,
		Action: runGenBhoversionRc,
	}
	genBhoversionRc.Flags.StringVar(&manifestJson, "manifest", manifestJson,
		"read version from manifest.json")
	ieExt.MustRegisterSubcommand(genBhoversionRc)

	getApp().MustRegisterSubcommand(ieExt)
}

var manifestJson string

// Subcommand handler.
func runGenBhoversionRc(cmd *gocli.Command, args []string) {
	if len(args) > 1 {
		cmd.Usage()
		os.Exit(2)
	}

	var filename string
	if len(args) == 1 {
		filename = args[0]
	} else {
		filename = "bhoversion.rc"
	}

	var version string
	if manifestJson != "" {
		var manifest struct {
			Version string
		}

		content, err := ioutil.ReadFile(manifestJson)
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}

		if err := json.Unmarshal(content, &manifest); err != nil {
			log.Fatalf("Error: %v\n", err)
		}

		version = manifest.Version
	}

	var ctx struct {
		CompanyName     string
		FileDescription string
		Version         string
		VersionCommas   string
		LegalCopyright  string
		ProductName     string
	}

	if err := nutrition.Env("BHRC_").Feed(&ctx); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if version != "" {
		ctx.Version = version
	}

	switch {
	case ctx.CompanyName == "":
		log.Fatalln("Error: BHRC_COMPANYNAME not set")
	case ctx.FileDescription == "":
		log.Fatalln("Error: BHRC_FILEDESCRIPTION not set")
	case ctx.Version == "":
		log.Fatalln("Error: BHRC_VERSION not set")
	case ctx.LegalCopyright == "":
		log.Fatalln("Error: BHRC_LEGALCOPYRIGHT not set")
	case ctx.ProductName == "":
		log.Fatalln("Error: BHRC_PRODUCTNAME not set")
	}

	buildNum := os.Getenv("BUILD_NUMBER")
	if buildNum == "" {
		buildNum = "0"
	}
	ctx.Version = fmt.Sprintf("%v.%v", ctx.Version, buildNum)

	matched, err := regexp.Match("[0-9]+([.][0-9]){3}", []byte(ctx.Version))
	if err != nil {
		panic(err)
	}
	if !matched {
		log.Fatalf("Error: invalid version string")
	}

	ctx.VersionCommas = strings.Replace(ctx.Version, ".", ",", -1)

	t := template.Must(template.New("bhoversion.rc").Parse(bhoversionRcTemplate))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if err := t.Execute(file, ctx); err != nil {
		file.Close()
		log.Fatalf("Error: %v\n", err)
	}

	if err := file.Close(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

const bhoversionRcTemplate = `
1 VERSIONINFO
FILEVERSION {{.VersionCommas}}
PRODUCTVERSION {{.VersionCommas}}
FILEOS 0x4
FILETYPE 0x2
{
BLOCK "StringFileInfo"
{
        BLOCK "040904b0"
        {
                VALUE "CompanyName", "{{.CompanyName}}"
                VALUE "FileDescription", "{{.FileDescription}}"
                VALUE "FileVersion", "{{.Version}}"
                VALUE "InternalName", "ancho.dll"
                VALUE "LegalCopyright", "{{.LegalCopyright}}"
                VALUE "OriginalFilename", "ancho.dll"
                VALUE "ProductName", "{{.ProductName}}"
                VALUE "ProductVersion", "{{.Version}}"
        }
}

BLOCK "VarFileInfo"
{
        VALUE "Translation", 0x0409 0x04E4
}
}
`
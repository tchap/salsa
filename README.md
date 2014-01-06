# Salsa - A Build Artifacts Manager

Salsa is a CLI utility for uploading build artifacts to a centralized artifact
store, mainly for sharing of artifacts between CI builds, but also for people
to be able to get a package by project name and version.

The main ideas behind Salsa are:

1. Be as platform-independent as possible.
2. Be as simple as possible.
3. Be compatible with NPM's `package.json`.

These points led to the decision of writing Salsa in Go and using a plain old
HTTP server with HTTP GET/PUT for download/upload of packages respectively.
There are absolutely no external dependencies and the program can be distributed
as a statically compiled binary.

## Project Status

This is very much in active development, not even all of the desired
subcommands are implemented yet.

## Installation

1. [Install Go](http://golang.org/doc/install)
2. [Set up a Go workspace](http://golang.org/doc/code.html) (make sure to add `bin/` to `PATH`)
3. `go get github.com/tchap/salsa`

## Usage

```
APPLICATION:
  salsa - a project build artifacts manager

USAGE:
  salsa [-v] [-dry] [-f CONFIG_FILE] [-username USER -password PASSWD] SUBCMD
  salsa [-v] [-dry] [-f CONFIG_FILE] [-username USER -ask_password] SUBCMD

VERSION:
  0.0.1

OPTIONS:
  -ask_password=false: ask for the Basic auth password
  -dry=false: just print what would be executed
  -f="package.json": config file to use
  -h=false: print help and exit
  -password="": Basic auth password
  -username="": Basic auth username
  -v=false: print verbose output

DESCRIPTION:
  Salsa can be used to upload or download project build artifacts to or from
  a remote HTTP server using HTTP GET/PUT respectively.

  Salsa can also use Basic auth to authenticate HTTP requests.

SUBCOMMANDS:
  publish	 publish build artifacts contained in a directory
  
```

### Subcommands

#### Publish

`salsa publish` reads `package.json` by default to get the project name and
version. The config file name can be overwritten by using a command line flag.

Then, Salsa searches the artifacts source directory for files and packs them
together into a single archive, which is then uploaded to the remote server
using HTTP PUT. Project name, version and build number is used to generate
the archive filename.

```
COMMAND:
  publish - publish build artifacts contained in a directory

USAGE:
  publish [-artifacts_dir DIR]
          [-archiver {tar|zip}]

  publish [-artifacts_dir DIR]
          [-archiver {tar|zip}] STORE_URL

  publish [-artifacts_dir DIR]
          [-archiver {tar|zip}] PKG_NAME PKG_VERSION STORE_URL

OPTIONS:
  -archiver="tar": archiver to use for packing of the artifacts
  -artifacts_dir="artifacts/publish": directory to be searched for artifacts
  -h=false: print help and exit

DESCRIPTION:
  This subcommand collects all artifacts contained in DIR, packs them up using
  the chosen archiver and uploads the archive to STORE_URL.

  The first form of this subscommand takes all necessary metadata from
  package.json (of any other config file that is set).

  The second form takes just PKG_NAME and PKG_VERSION from the config file,
  STORE_URL is specified on the command line.

  The last form requires no config file since everything is specified manually.

ENVIRONMENTAL VARIABLES:
  BUILD_NUMBER - if set, the version is set to <version>.${BUILD_NUM}
		
```

## Example

Check the `example` directory for a life demo.

## TODO:

* Handle interrupts.

## License

MIT, can be found in the LICENSE file.

# Salsa - A Build Artifacts Manager

Salsa is a CLI utility for uploading build artifacts to a centralized artifact
store, mainly for sharing of artifacts between Continuous Integration builds,
but also for people to be able to get a package by project name and version.

The main ideas behind Salsa are:

1. Be as platform-independent as possible.
2. Be as simple as possible.
3. Be somehow compatible with NPM's `package.json`.

These points led to the decision of writing Salsa in Go and using a plain old
HTTP server with HTTP GET/PUT for download/upload of packages respectively.
There are absolutely no external dependencies and the program can be distributed
as a statically linked binary.

## Project Status

This is very much in active development, not all subcommands are implemented
yet.

## Installation

1. [Install Go](http://golang.org/doc/install)
2. [Set up a Go workspace](http://golang.org/doc/code.html) (make sure to add `bin/` to `PATH`)
3. `go get github.com/tchap/salsa`
4. PROFIT!

## Usage

```
APPLICATION:
  salsa - a project build artifacts manager

USAGE:
  salsa [-h] [-v] [-dry] [-username USER -password PASSWD] SUBCMD

VERSION:
  0.0.1

OPTIONS:
  -dry=false: just print what would be executed
  -h=false: print help and exit
  -password="": Basic auth password
  -username="": Basic auth username
  -v=false: print verbose output

DESCRIPTION:
  Salsa is a project build artifacts manager that can publish or fetch build
  artifacts. A remote HTTP server acts as the artifacts store and salsa uses
  HTTP PUT and GET requests to publish and fetch the artifacts respectively.
  See the subcommands for more details.

  Salsa can be set up to use Basic authentication to authenticate HTTP requests.
  If you, however, do not want to specify the credentials on the command line,
  $HOME/.salsarc can be used to set them for you.

ENVIRONMENTAL VARIABLES:
  SALSA_USER_CONFIG - overwrites the default location for the user-specific
                      configuration file, which is $HOME/.salsarc

SUBCOMMANDS:
  publish	 publish build artifacts
  
```

### Subcommands

#### Publish

```
COMMAND:
  publish - publish build artifacts

USAGE:
  publish [-tag TAG] [-archiver {tar|zip}] ARTIFACTS_DIR

OPTIONS:
  -archiver="tar": archiver to use for packing the artifacts
  -h=false: print help and exit
  -tag="": tag to use in the archive file name

DESCRIPTION:
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
      Project URL secrets are also stored there under "secrets.$project"

ENVIRONMENTAL VARIABLES:
  BRANCH       - if set, $BRANCH is used in the archive filename as $branch
  BUILD_NUMBER - if set, $version is set to $version.$BUILD_NUMBER
		
```

## Example

Check the `example` directory for a life demo.

## TODO

* Handle interrupts.

## License

MIT, can be found in the LICENSE file.

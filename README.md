# xsubfind3r

![made with go](https://img.shields.io/badge/made%20with-Go-0000FF.svg) [![release](https://img.shields.io/github/release/hueristiq/xsubfind3r?style=flat&color=0000FF)](https://github.com/hueristiq/xsubfind3r/releases) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=0000FF)](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0000FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/xsubfind3r.svg?style=flat&color=0000FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/xsubfind3r.svg?style=flat&color=0000FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-0000FF.svg)](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md)

`xsubfind3r` is a command-line interface (CLI) utility to find domain's known subdomains passively.

## Resource

* [Features](#features)
* [Installation](#installation)
	* [Install release binaries](#install-release-binaries)
	* [Install source](#install-sources)
		* [`go install ...`](#go-install)
		* [`go build ...` the development Version](#go-build--the-development-version)
* [Post Installation](#post-installation)
* [Usage](#usage)
* [Contribution](#contribution)
* [Licensing](#licensing)

## Features

* Curated list of 21 passive sources to maximize results
* Optimized for speed and lightweight on resources

## Installation

### Install release binaries

Visit the [releases page](https://github.com/hueristiq/xsubfind3r/releases) and find the appropriate archive for your operating system and architecture. Download the archive from your browser or copy its URL and retrieve it with `wget` or `curl`:

* ...with `wget`:

	```bash
	wget https://github.com/hueristiq/xsubfind3r/releases/download/v<version>/xsubfind3r-<version>-linux-amd64.tar.gz
	```

* ...or, with `curl`:

	```bash
	curl -OL https://github.com/hueristiq/xsubfind3r/releases/download/v<version>/xsubfind3r-<version>-linux-amd64.tar.gz
	```

...then, extract the binary:

```bash
tar xf xsubfind3r-<version>-linux-amd64.tar.gz
```

> **TIP:** The above steps, download and extract, can be combined into a single step with this onliner
> 
> ```bash
> curl -sL https://github.com/hueristiq/xsubfind3r/releases/download/v<version>/xsubfind3r-<version>-linux-amd64.tar.gz | tar -xzv
> ```

**NOTE:** On Windows systems, you should be able to double-click the zip archive to extract the `xsubfind3r` executable.

...move the `xsubfind3r` binary to somewhere in your `PATH`. For example, on GNU/Linux and OS X systems:

```bash
sudo mv xsubfind3r /usr/local/bin/
```

**NOTE:** Windows users can follow [How to: Add Tool Locations to the PATH Environment Variable](https://msdn.microsoft.com/en-us/library/office/ee537574(v=office.14).aspx) in order to add `xsubfind3r` to their `PATH`.

### Install source

Before you install from source, you need to make sure that Go is installed on your system. You can install Go by following the official instructions for your operating system. For this, we will assume that Go is already installed.

#### `go install ...`

```bash
go install -v github.com/hueristiq/xsubfind3r/cmd/xsubfind3r@latest
```

#### `go build ...` the development Version

* Clone the repository

	```bash
	git clone https://github.com/hueristiq/xsubfind3r.git 
	```

* Build the utility

	```bash
	cd xsubfind3r/cmd/xsubfind3r && \
	go build .
	```

* Move the `xsubfind3r` binary to somewhere in your `PATH`. For example, on GNU/Linux and OS X systems:

	```bash
	sudo mv xsubfind3r /usr/local/bin/
	```

	**NOTE:** Windows users can follow [How to: Add Tool Locations to the PATH Environment Variable](https://msdn.microsoft.com/en-us/library/office/ee537574(v=office.14).aspx) in order to add `xsubfind3r` to their `PATH`.


**NOTE:** While the development version is a good way to take a peek at `xsubfind3r`'s latest features before they get released, be aware that it may have bugs. Officially released versions will generally be more stable.

## Post Installation

`xsubfind3r` will work after [installation](#installation). However, to configure `xsubfind3r` to work with certain sources will require API keys to work. Currently these services include:

* chaos
* github
* intelx

The API keys are stored in the `$HOME/.hueristiq/xurlfind3r/config.yaml` file - created upon first run - and uses the YAML format.

Example:

```yaml
version: 1.4.0
sources:
    - alienvault
    - anubis
    - archiveis
    - bufferover
    - cebaidu
    - certspotterv0
    - chaos
    - crtsh
    - github
    - hackertarget
    - intelx
    - rapiddns
    - riddler
    - sonar
    - sublist3r
    - threatcrowd
    - threatminer
    - urlscan
    - wayback
    - ximcx
keys:
    chaos:
        - d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39asdsd54bbc1aabb208c9acfb
    github:
        - d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39
        - asdsd54bbc1aabb208c9acfbd2dd41ce7fc9db39
    intelx:
        - 2.intelx.io:00000000-0000-0000-0000-000000000000
```

## Usage

To display help message for `xsubfind3r` use the `-h` flag:

```bash
xsubfind3r -h
```

help message:

```text
                _      __ _           _ _____      
__  _____ _   _| |__  / _(_)_ __   __| |___ / _ __ 
\ \/ / __| | | | '_ \| |_| | '_ \ / _` | |_ \| '__|
 >  <\__ \ |_| | |_) |  _| | | | | (_| |___) | |   
/_/\_\___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_| v0.0.0

A CLI utility to find domain's known subdomains.

USAGE:
  xsubfind3r [OPTIONS]

TARGET:
  -d, --domain string             target domain

SOURCES:
 -s,  --sources bool              list available sources
      --exclude-sources strings   comma(,) separated list of sources to exclude
      --use-sources strings       comma(,) separated list of sources to use

OUTPUT:
  -m, --monochrome                no colored output mode
  -o, --output string             output file to write found URLs
  -v, --verbosity                 debug, info, warning, error, fatal or silent (default: info)
```

## Contribution

[Issues](https://github.com/hueristiq/xsubfind3r/issues) and [Pull Requests](https://github.com/hueristiq/xsubfind3r/pulls) are welcome! Check out the [contribution guidelines](./CONTRIBUTING.md).

## Licensing

This utility is distributed under the [MIT license](./LICENSE).
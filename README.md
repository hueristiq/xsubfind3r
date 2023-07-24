# xsubfind3r

![made with go](https://img.shields.io/badge/made%20with-Go-0000FF.svg) [![release](https://img.shields.io/github/release/hueristiq/xsubfind3r?style=flat&color=0000FF)](https://github.com/hueristiq/xsubfind3r/releases) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=0000FF)](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0000FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/xsubfind3r.svg?style=flat&color=0000FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/xsubfind3r.svg?style=flat&color=0000FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-0000FF.svg)](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md)

`xsubfind3r` is a command-line interface (CLI) utility to find domain's known subdomains from curated passive online sources.

## Resource

* [Features](#features)
* [Installation](#installation)
	* [Install release binaries (Without Go Installed)](#install-release-binaries-without-go-installed)
	* [Install source (With Go Installed)](#install-source-with-go-installed)
		* [`go install ...`](#go-install)
		* [`go build ...` the development Version](#go-build--the-development-version)
* [Post Installation](#post-installation)
* [Usage](#usage)
* [Contribution](#contribution)
* [Licensing](#licensing)

## Features

* Fetches domains from curated passive sources to maximize results.
* Supports `stdin` and `stdout` for easy integration into workflows.
* Cross-Platform (Windows, Linux & macOS).

## Installation

### Install release binaries (Without Go Installed)

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

### Install source (With Go Installed)

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

`xsubfind3r` will work right after [installation](#installation). However, **[BeVigil](https://bevigil.com)**, **[Chaos](https://chaos.projectdiscovery.io/#/)**, **[Fullhunt](https://fullhunt.io/)**, **[Github](https://github.com)**, **[Intelligence X](https://intelx.io)** and **[Shodan](https://shodan.io/)** require API keys to work, **[URLScan](https://urlscan.io)** supports API key but not required. The API keys are stored in the `$HOME/.hueristiq/xsubfind3r/config.yaml` file - created upon first run - and uses the YAML format. Multiple API keys can be specified for each of these source from which one of them will be used.

Example `config.yaml`:

```yaml
version: 0.2.0
sources:
    - alienvault
    - anubis
    - bevigil
    - chaos
    - commoncrawl
    - crtsh
    - fullhunt
    - github
    - hackertarget
    - intelx
    - shodan
    - urlscan
    - wayback
keys:
    bevigil:
        - awA5nvpKU3N8ygkZ
    chaos:
        - d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39asdsd54bbc1aabb208c9acfb
    fullhunt:
        - 0d9652ce-516c-4315-b589-9b241ee6dc24
    github:
        - d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39
        - asdsd54bbc1aabb208c9acfbd2dd41ce7fc9db39
    intelx:
        - 2.intelx.io:00000000-0000-0000-0000-000000000000
    shodan:
        - AAAAClP1bJJSRMEYJazgwhJKrggRwKA
    urlscan:
        - d4c85d34-e425-446e-d4ab-f5a3412acbe8
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
/_/\_\___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_| v0.2.0

USAGE:
  xsubfind3r [OPTIONS]

INPUT:
 -d, --domain string[]                 target domains
 -l, --list string                     target domains' list file path

SOURCES:
      --sources bool                   list supported sources
 -u,  --sources-to-use string[]        comma(,) separeted sources to use
 -e,  --sources-to-exclude string[]    comma(,) separeted sources to exclude

OPTIMIZATION:
 -t,  --threads int                    number of threads (default: 50)

OUTPUT:
     --no-color bool                   disable colored output
 -o, --output string                   output subdomains' file path
 -O, --output-directory string         output subdomains' directory path
 -v, --verbosity string                debug, info, warning, error, fatal or silent (default: info)

CONFIGURATION:
 -c,  --configuration string           configuration file path (default: ~/.hueristiq/xsubfind3r/config.yaml)
```

## Contribution

[Issues](https://github.com/hueristiq/xsubfind3r/issues) and [Pull Requests](https://github.com/hueristiq/xsubfind3r/pulls) are welcome! **Check out the [contribution guidelines](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md).**

## Licensing

This utility is distributed under the [MIT license](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE).
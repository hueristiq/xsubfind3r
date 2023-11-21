# xsubfind3r

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go report card](https://goreportcard.com/badge/github.com/hueristiq/xsubfind3r)](https://goreportcard.com/report/github.com/hueristiq/xsubfind3r) [![release](https://img.shields.io/github/release/hueristiq/xsubfind3r?style=flat&color=1E90FF)](https://github.com/hueristiq/xsubfind3r/releases) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/xsubfind3r.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/xsubfind3r.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:closed) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md)

`xsubfind3r` is a command-line interface (CLI) based passive subdomain discovery utility. It is designed to efficiently identify known subdomains of given domains by tapping into a multitude of curated online passive sources.

## Resource

* [Features](#features)
* [Installation](#installation)
	* [Install release binaries (Without Go Installed)](#install-release-binaries-without-go-installed)
	* [Install source (With Go Installed)](#install-source-with-go-installed)
		* [`go install ...`](#go-install)
		* [`go build ...` the development Version](#go-build--the-development-version)
* [Post Installation](#post-installation)
* [Usage](#usage)
* [Contributing](#contributing)
* [Licensing](#licensing)
* [Credits](#credits)
    * [Contributors](#contributors)
    * [Similar Projects](#similar-projects)

## Features

* Fetches subdomains from curated passive sources to maximize results.
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

`xsubfind3r` will work right after [installation](#installation). However, **[BeVigil](https://bevigil.com)**, **[BufferOver](https://tls.bufferover.run/)**, **[BuiltWith](https://api.builtwith.com/domain-api)**, **[Certspotter](https://sslmate.com/ct_search_api/)**, **[Chaos](https://chaos.hueristiq.io/#/)**, **[Fullhunt](https://fullhunt.io/)**, **[Github](https://github.com)**, **[Intelligence X](https://intelx.io)**, **[LeakIX](https://leakix.net)** and **[Shodan](https://shodan.io/)** require API keys to work, **[URLScan](https://urlscan.io)** supports API key but not required. The API keys are stored in the `$HOME/.config/xsubfind3r/config.yaml` file - created upon first run - and uses the YAML format. Multiple API keys can be specified for each of these source from which one of them will be used.

Example `config.yaml`:

> **NOTE:** The keys/tokens below are invalid and used as examples, use your own keys/tokens!

```yaml
version: 0.5.0
sources:
    - alienvault
    - anubis
    - bevigil
    - bufferover
    - builtwith
    - certspotter
    - chaos
    - commoncrawl
    - crtsh
    - fullhunt
    - github
    - hackertarget
    - intelx
    - leakix
    - shodan
    - urlscan
    - wayback
keys:
    bevigil:
        - awA5nvpKU3N8ygkZ
    bufferover:
        - COx9GBnhz63hcF1hlBtLb4KAdlzJly1d8xeovTjK
    builtwith:
        - 7fcbaec4-dc49-472c-b837-3896cb255823
    chaos:
        - d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39asdsd54bbc1aabb208c9acfb
    fullhunt:
        - 0d9652ce-516c-4315-b589-9b241ee6dc24
    github:
        - d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39
        - asdsd54bbc1aabb208c9acfbd2dd41ce7fc9db39
    intelx:
        - 2.intelx.io:00000000-0000-0000-0000-000000000000
    leakix:
        - xhDsgKejYTUWVNLn9R6f8afhsG6h6KM69lqEBoMJbfcvDk1v
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
/_/\_\___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_|
                                             v0.5.0

                   with <3 by Hueristiq Open Source

USAGE:
 xsubfind3r [OPTIONS]

CONFIGURATION:
 -c, --configuration string            configuration file (default: $HOME/.config/xsubfind3r/config.yaml)

INPUT:
 -d, --domain string[]                 target domain
 -l, --list string                     target domains list file path

TIP: For multiple input domains use comma(,) separated value with `-d`,
     specify multiple `-d`, load from file with `-l` or load from stdin.

SOURCES:
     --sources bool                    list supported sources
 -u, --sources-to-use string[]         comma(,) separated sources to use
 -e, --sources-to-exclude string[]     comma(,) separated sources to exclude

OUTPUT:
     --monochrome bool                 display no color output
 -o, --output string                   output subdomains file path
 -O, --output-directory string         output subdomains directory path
 -s, --silent bool                     display output subdomains only
 -v, --verbose bool                    display verbose output
```

## Contributing

[Issues](https://github.com/hueristiq/xsubfind3r/issues) and [Pull Requests](https://github.com/hueristiq/xsubfind3r/pulls) are welcome! **Check out the [contribution guidelines](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md).**

## Licensing

This utility is distributed under the [MIT license](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE).

## Credits

### Contributors

Thanks to the amazing [contributors](https://github.com/hueristiq/xsubfind3r/graphs/contributors) for keeping this project alive.

[![contributors](https://contrib.rocks/image?repo=hueristiq/xsubfind3r&max=500)](https://github.com/hueristiq/xsubfind3r/graphs/contributors)

### Similar Projects

Thanks to similar open source projects - check them out, may fit in your workflow.

[subfinder](https://github.com/projectdiscovery/subfinder) ◇ [assetfinder](https://github.com/tomnomnom/assetfinder)
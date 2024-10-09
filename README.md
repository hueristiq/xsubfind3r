# xsubfind3r

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go report card](https://goreportcard.com/badge/github.com/hueristiq/xsubfind3r)](https://goreportcard.com/report/github.com/hueristiq/xsubfind3r) [![release](https://img.shields.io/github/release/hueristiq/xsubfind3r?style=flat&color=1E90FF)](https://github.com/hueristiq/xsubfind3r/releases) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/xsubfind3r.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/xsubfind3r.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/xsubfind3r/issues?q=is:issue+is:closed) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md)

`xsubfind3r` is a command-line utility designed to help you discover subdomains for a given domain in a simple, efficient way. It works by gathering information from a variety of passive sources, meaning it doesn't interact directly with the target but instead gathers data that is already publicly available.

The utility employs several techniques for subdomain discovery, including:

* **Certificate Transparency Logs** to identify subdomains through public certificate databases. (`certspotter`, `crtsh`, `certificatedetails`)
* **Search Engine Queries** to leverage public repositories and services that index domains. (`github`, `hackertarget`, `securitytrails`)
* **DNS-based Enumeration** to gather information from various DNS resolvers and registrars. (`bevigil`, `chaos`, `shodan`, `censys`)
* **Passive DNS databases** to retrieve records of domains and subdomains. (`securitytrails`, `censys`, `shodan`)
* **Web Archives and Historical Data** to find past subdomain references through web archives. (`wayback`, `commoncrawl`)
* **Public Code Repositories** to search for subdomains in open-source projects. (`github`)
* **Threat Intelligence Feeds** to query platforms for known subdomains and associated intelligence.  (`otx`, `fullhunt`, `leakix`, `intelx`)
* **OSINT Platforms** to collect publicly known subdomains from various open-source intelligence sources. (`anubis`, `fullhunt`, `bevigil`, `urlscan`)
* **Company Profiling and Technology Stacks** to reveal subdomains associated with specific technologies used by companies. (`builtwith`)
* **Miscellaneous**. (`subdomaincenter`)

This makes `xsubfind3r` a powerful tool for security researchers, IT professionals, and anyone looking to gain insights into the subdomains associated with a domain.

## Resource

* [Features](#features)
* [Installation](#installation)
	* [Install release binaries (Without Go Installed)](#install-release-binaries-without-go-installed)
	* [Install on Docker (With Docker Installed)](#install-on-docker-with-docker-installed)
	* [Install source (With Go Installed)](#install-source-with-go-installed)
		* [`go install ...`](#go-install)
		* [`go build ...`](#go-build)
* [Post Installation](#post-installation)
* [Usage](#usage)
* [Contributing](#contributing)
* [Licensing](#licensing)
* [Credits](#credits)
    * [Contributors](#contributors)
    * [Similar Projects](#similar-projects)

## Features

* **Wide Coverage:** Fetches subdomains from multiple online passive sources to provide extensive results.
* **Flexible Output:** Supports silent mode for only showing results and offers various output options for saving the results.
* **Easy to Integrate:** Supports `stdin` and `stdout` for easy integration in automated workflows.
* **Cross-Platform:** Works on Windows, Linux, and macOS.

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

> [!TIP]
> The above steps, download and extract, can be combined into a single step with this onliner
> 
> ```bash
> curl -sL https://github.com/hueristiq/xsubfind3r/releases/download/v<version>/xsubfind3r-<version>-linux-amd64.tar.gz | tar -xzv
> ```

> [!NOTE]
> On Windows systems, you should be able to double-click the zip archive to extract the `xsubfind3r` executable.

...move the `xsubfind3r` binary to somewhere in your `PATH`. For example, on GNU/Linux and OS X systems:

```bash
sudo mv xsubfind3r /usr/local/bin/
```

> [!NOTE]
> Windows users can follow [How to: Add Tool Locations to the PATH Environment Variable](https://msdn.microsoft.com/en-us/library/office/ee537574(v=office.14).aspx) in order to add `xsubfind3r` to their `PATH`.

### Install on Docker (With Docker Installed)

If you have Docker installed, you can use `xsubfind3r` using it's image:

* Pull the docker image using:

    ```bash
    docker pull hueristiq/xsubfind3r:latest
    ```

* Run `xsubfind3r` using the image:

    ```bash
    docker run --rm hueristiq/xsubfind3r:latest -h
    ```

### Install source (With Go Installed)

Before you install from source, you need to make sure that Go is installed on your system. You can install Go by following the [official instructions](https://go.dev/doc/install) for your operating system. For this, we will assume that Go is already installed.

#### `go install ...`

```bash
go install -v github.com/hueristiq/xsubfind3r/cmd/xsubfind3r@latest
```

#### `go build ...`

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

	Windows users can follow [How to: Add Tool Locations to the PATH Environment Variable](https://msdn.microsoft.com/en-us/library/office/ee537574(v=office.14).aspx) in order to add `xsubfind3r` to their `PATH`.


> [!CAUTION]
> While the development version is a good way to take a peek at `xsubfind3r`'s latest features before they get released, be aware that it may have bugs. Officially released versions will generally be more stable.

## Post Installation

`xsubfind3r` will work right after [installation](#installation). However, **[BeVigil](https://bevigil.com)**, **[BuiltWith](https://api.builtwith.com/domain-api)**, **[Censys](https://censys.com/)**, **[Certspotter](https://sslmate.com/ct_search_api/)**, **[Chaos](https://chaos.hueristiq.io/#/)**, **[Fullhunt](https://fullhunt.io/)**, **[Github](https://github.com)**, **[Intelligence X](https://intelx.io)**, **[LeakIX](https://leakix.net)**, **[Security Trails](https://securitytrails.com/)**, **[Shodan](https://shodan.io/)** and **[VirusTotal](https://www.virustotal.com)** require API keys to work, **[URLScan](https://urlscan.io)** supports API key but not required. The API keys are stored in the `$HOME/.config/xsubfind3r/config.yaml` file, created upon first run, and uses the YAML format, or supplied via environment variables. Multiple API keys can be specified for each of these source from which one of them will be used.

Example `config.yaml`:

> [!CAUTION]
> The keys/tokens below are invalid and used as examples, use your own keys/tokens!

```yaml
version: 0.8.0
sources:
    - anubis
    - bevigil
    - builtwith
    - censys
    - certificatedetails
    - certspotter
    - chaos
    - commoncrawl
    - crtsh
    - fullhunt
    - github
    - hackertarget
    - intelx
    - leakix
    - otx
    - securitytrails
    - shodan
    - subdomaincenter
    - urlscan
    - wayback
    - virustotal
keys:
    bevigil:
        - awA5nvpKU3N8ygkZ
    builtwith:
        - 7fcbaec4-dc49-472c-b837-3896cb255823
    censys:
        - 0d9652ce-516c-4315-b589-9b241ee6dc24:AAAAClP1bJJSRMEYJazgwhJKrggRwKA
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
    securitytrails:
        - smiSOZcrtVI214MfkLb6FoCFqXGEhcdG
    shodan:
        - AAAAClP1bJJSRMEYJazgwhJKrggRwKA
    urlscan:
        - d4c85d34-e425-446e-d4ab-f5a3412acbe8
    virustotal:
        - 088d5d8afdfd9ac22388e9ebb3cc60e14g92bcdf2f80680d0938116139499410
```

> [!NOTE]
> To run `xsubfind3r` using docker with a local config file:
>
>```bash
>docker run --rm -v $HOME/.config/xsubfind3r:/root/.config/xsubfind3r hueristiq/xsubfind3r:latest -h
>```

Example environmet variables:

```bash
XURLFIND3R_KEYS_BEVIGIL=awA5nvpKU3N8ygkZ
XURLFIND3R_KEYS_BUILTWITH=7fcbaec4-dc49-472c-b837-3896cb255823
XURLFIND3R_KEYS_CHAOS=d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39asdsd54bbc1aabb208c9acfb
XURLFIND3R_KEYS_FULLHUNT=0d9652ce-516c-4315-b589-9b241ee6dc24
XURLFIND3R_KEYS_GITHUB=asdsd54bbc1aabb208c9acfbd2dd41ce7fc9db39,d23a554bbc1aabb208c9acfbd2dd41ce7fc9db39
XURLFIND3R_KEYS_INTELX=2.intelx.io:00000000-0000-0000-0000-000000000000
XURLFIND3R_KEYS_LEAKIX=xhDsgKejYTUWVNLn9R6f8afhsG6h6KM69lqEBoMJbfcvDk1v
XURLFIND3R_KEYS_SHODAN=AAAAClP1bJJSRMEYJazgwhJKrggRwKA
XURLFIND3R_KEYS_URLSCAN=d4c85d34-e425-446e-d4ab-f5a3412acbe8
```

## Usage

To start using `xsubfind3r`, open your terminal and run the following command for a list of options:

```bash
xsubfind3r -h
```

Here's what the help message looks like:

```text
                _      __ _           _ _____
__  _____ _   _| |__  / _(_)_ __   __| |___ / _ __
\ \/ / __| | | | '_ \| |_| | '_ \ / _` | |_ \| '__|
 >  <\__ \ |_| | |_) |  _| | | | | (_| |___) | |
/_/\_\___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_|
                                             v0.8.0

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

For example, to discover subdomains for `example.com`:

```bash
xsubfind3r -d example.com
```

You can also use multiple domains by separating them with commas or providing a list from a file.

## Contributing

We welcome contributions! Feel free to submit [Pull Requests](https://github.com/hueristiq/xsubfind3r/pulls) or report [Issues](https://github.com/hueristiq/xsubfind3r/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/xsubfind3r/blob/master/CONTRIBUTING.md).

## Licensing

This utility is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/xsubfind3r/blob/master/LICENSE).

## Credits

### Contributors

A huge thanks to all the contributors who have helped make `xsubfind3r` what it is today!

[![contributors](https://contrib.rocks/image?repo=hueristiq/xsubfind3r&max=500)](https://github.com/hueristiq/xsubfind3r/graphs/contributors)

### Similar Projects

If you're interested in more utilities like this, check out:

[subfinder](https://github.com/projectdiscovery/subfinder) ◇ [assetfinder](https://github.com/tomnomnom/assetfinder)
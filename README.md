# subfind3r

[![release](https://img.shields.io/github/release/hueristiq/subfind3r?style=flat&color=0040ff)](https://github.com/hueristiq/subfind3r/releases) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0040ff.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/subfind3r.svg?style=flat&color=0040ff)](https://github.com/hueristiq/subfind3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/subfind3r.svg?style=flat&color=0040ff)](https://github.com/hueristiq/subfind3r/issues?q=is:issue+is:closed) [![license](https://img.shields.io/badge/license-MIT-gray.svg?colorB=0040FF)](https://github.com/hueristiq/subfind3r/blob/master/LICENSE) [![twitter](https://img.shields.io/badge/twitter-@itshueristiq-0040ff.svg)](https://twitter.com/itshueristiq)

subfind3r is a passive subdomain discovery tool - it gathers a list of subdomains passively using a curated list of passive online sources.

## Usage

To display help message for subfind3r use the `-h` flag:

```bash
subfind3r -h
```

```text
           _      __ _           _ _____      
 ___ _   _| |__  / _(_)_ __   __| |___ / _ __ 
/ __| | | | '_ \| |_| | '_ \ / _` | |_ \| '__|
\__ \ |_| | |_) |  _| | | | | (_| |___) | |   
|___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_| v1.4.0

USAGE:
  subfind3r [OPTIONS]

OPTIONS:
  -d,  --domain            domain to find subdomains for
  -eS, --exclude-sources   comma(,) separated list of sources to exclude
  -lS, --list-sources      list all the sources available
  -nC, --no-color          no color mode: Don't use colors in output
  -s,  --silent            silent mode: Output subdomains only
  -uS, --use-sources       comma(,) separated list of sources to use
```

**DESCLAIMER:** wayback and github sources are a bit slow.

## Installation

#### From Binary

You can download the pre-built binary for your platform from this repository's [releases](https://github.com/hueristiq/subfind3r/releases/) page, extract, then move it to your `$PATH`and you're ready to go.

#### From Source

subfind3r requires **go1.17+** to install successfully. Run the following command to get the repo:-

```bash
go install -v github.com/hueristiq/subfind3r/cmd/subfind3r@latest
```

#### From Github

```bash
git clone https://github.com/hueristiq/subfind3r.git && \
cd subfind3r/cmd/subfind3r/ && \
go build . && \
mv subfind3r /usr/local/bin/ && \
subfind3r -h
```

## Post Installation

subfind3r will work after [installation](#installation). However, to configure subfind3r to work with certain services you will need to have setup API keys. Currently these services include:

* chaos
* github
* intelx

The API keys are stored in the `$HOME/.config/subfind3r/conf.yaml` file - created upon first run - and uses the YAML format. Multiple API keys can be specified for each of these services.

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
## Contribution

[Issues](https://github.com/hueristiq/subfind3r/issues) and [Pull Requests](https://github.com/hueristiq/subfind3r/pulls) are welcome! 
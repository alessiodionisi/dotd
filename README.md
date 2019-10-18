# dotd

[![Release](https://img.shields.io/github/release/adnsio/dotd?style=flat-square)](https://github.com/adnsio/dotd/releases/latest)
[![License](https://img.shields.io/github/license/adnsio/dotd?style=flat-square)](https://github.com/adnsio/dotd/blob/master/LICENSE)

Local proxy to DNS over HTTPS (RFC 8484) compatible servers.

## Installation

`dotd` supports macOS, Linux and Windows.

### Binaries

We distribute `dotd` binaries on [every release](https://github.com/adnsio/dotd/releases).

### Homebrew on macOS

```
$ brew install adnsio/tap/dotd
```

## How to use

Simply run `dotd` executable then set `127.0.0.1` for IPv4 and `::1` for IPv6 as primary DNS server.

## Options

```
Usage: dotd [options]
  -address string
        udp address (default "[::]:53")
  -logs
        enable logs
  -upstream string
        upstream dns server (default "https://1.1.1.1/dns-query")
  -version
        output version
```

## License

`dotd` is made with ♥ by [contributors](https://github.com/adnsio/dotd/graphs/contributors) and it's released under the GNU GPLv3 license.

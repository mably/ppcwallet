ppcwallet
=========

[![Build Status](https://travis-ci.org/mably/ppcwallet.png?branch=master)]
(https://travis-ci.org/mably/ppcwallet)
[![tip for next commit](http://peer4commit.com/projects/130.svg)](http://peer4commit.com/projects/130)

ppcwallet is a daemon handling bitcoin wallet functionality for a
single user.  It acts as both an RPC client to ppcd and an RPC server
for wallet clients and legacy RPC applications.  It is based on the
Conformal [btcwallet](https://github.com/conformal/btcwallet) code.

The wallet file format is based on
[Armory](https://github.com/etotheipi/BitcoinArmory) and provides a
deterministic wallet where all future generated private keys can be
recovered from a previous wallet backup.  Unencrypted wallets are
unsupported and are never written to disk.  This design decision has
the consequence of generating new wallets on the fly impossible: a
client is required to provide a wallet encryption passphrase.

ppcwallet is not an SPV client and requires connecting to a local or
remote ppcd instance for asynchronous blockchain queries and
notifications over websockets.  Full ppcd installation instructions
can be found [here](https://github.com/mably/ppcd).

As a daemon, ppcwallet provides no user interface and an additional
graphical or command line client is required for normal, personal
wallet usage.  Conformal has written
[btcgui](https://github.com/conformal/btcgui) as a graphical client
to btcwallet. It has not been ported to ppcwallet yet.

This project is currently under active development is not production
ready yet.  Support for creating and using wallets the main Bitcoin
network is currently disabled by default.

## Installation

### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Run the following commands to obtain and install ppcwallet and all
  dependencies:
```bash
$ go get -u -v github.com/mably/ppcd/...
$ go get -u -v github.com/mably/ppcwallet/...
```

- ppcd and ppcwallet will now be installed in either ```$GOROOT/bin``` or
  ```$GOPATH/bin``` depending on your configuration.  If you did not already
  add to your system path during the installation, we recommend you do so now.

## Updating

### Build from Source

- Run the following commands to update ppcwallet, all dependencies, and install it:

```bash
$ go get -u -v github.com/mably/ppcd/...
$ go get -u -v github.com/mably/ppcwallet/...
```

## Getting Started

The follow instructions detail how to get started with ppcwallet
connecting to a localhost ppcd.

### Windows

### Linux/BSD/POSIX/Source

- Run the following command to start ppcd:

```bash
$ ppcd --testnet -u rpcuser -P rpcpass
```

- Run the following command to start ppcwallet:

```bash
$ ppcwallet -u rpcuser -P rpcpass
```

If everything appears to be working, it is recommended at this point to
copy the sample ppcd and ppcwallet configurations and update with your
RPC username and password.

```bash
$ cp $GOPATH/src/github.com/mably/ppcd/sample-btcd.conf ~/.ppcd/ppcd.conf
$ cp $GOPATH/src/github.com/mably/ppcwallet/sample-btcwallet.conf ~/.ppcwallet/ppcwallet.conf
$ $EDITOR ~/.ppcd/ppcd.conf
$ $EDITOR ~/.ppcwallet/ppcwallet.conf
```

## Client Usage

Clients wishing to use ppcwallet must connect to the `ws` endpoint
over a websocket connection.  Messages sent to ppcwallet over this
websocket are expected to follow the standard Bitcoin JSON API
(partially documented
[here](https://en.bitcoin.it/wiki/Original_Bitcoin_client/API_Calls_list)).
Websocket connections also enable additional API extensions and
JSON-RPC notifications (currently undocumented).  The ppcd packages
`btcjson` and `btcws` provide types and functions for creating and
JSON (un)marshaling these requests and notifications.

## Issue Tracker

The [integrated github issue tracker](https://github.com/mably/ppcwallet/issues)
is used for this project.

<!--## GPG Verification Key

All official release tags are signed by Conformal so users can ensure the code
has not been tampered with and is coming from Conformal.  To verify the
signature perform the following:

- Download the public key from the Conformal website at
  https://opensource.conformal.com/GIT-GPG-KEY-conformal.txt

- Import the public key into your GPG keyring:
  ```bash
  gpg --import GIT-GPG-KEY-conformal.txt
  ```

- Verify the release tag with the following command where `TAG_NAME` is a
  placeholder for the specific tag:
  ```bash
  git tag -v TAG_NAME
  ```
-->
## License

ppcwallet is licensed under the liberal ISC License.

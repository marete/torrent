# torrent

`torrent` is a simple command-line BitTorrent client based on a
straight-forward usage of the excellent
[https://github.com/anacrolix/torrent](github.com/anacrolix/torrent)
library.

## Goals

The emphasis for torrent is simplicity and it is supposed to always
remain a simple command-line program with reasonable defaults.

I have only tested this on Linux, but there is no reason it shouldn't
run well on any other platform supported by Go.

## Usage

    torrent -magnet '<magnet link>'
    torrent -magnet '<magnet link>' -dir <dest_dir>
    torrent -file   '<path to .torrent file'
    torrent -download_bandwidth_limit 4 -upload_bandwidth_limit 1 -magnet '<magnet link>'

The download and upload limits are in Mbits/s (note: Bits *not* Bytes,
and M is for *Mega).
    
By default `torrent` writes the downloaded file(s) in the user's
`Downloads` directory. This can be changed with the `-dir`
command-line option.

## Features

`torrent` supports much of what the *anacrolix* package supports and
can otherwise be trivially extended to support other features of that
library. Perhaps the most notable features are:

* Support for continuing interrupted downloads.
* Support for download and upload limits.
* Support for both magnet links and torrent files.
* Support for multiple instances running on the same machine, downloading different torrents.

# Installation

    go install -v github.com/marete/torrent@latest

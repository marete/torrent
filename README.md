torrent is a simple command-line BitTorrent client based on a
straight-forward usage of the excellent
[https://github.com/anacrolix/torrent](github.com/anacrolix/torrent)
library.

Usage:

    torrent -magnet '<magnet link>'
    torrent -magnet '<magnet link>' -dir <dest_dir>
    torrent -file   '<path to .torrent file'
    
By default `torrent` writes the downloaded file(s) in the user's
`Downloads` directory. This can be changed with the `-dir`
command-line option.

`torrent` supports continuing interrupted downloads.

Installation

    go install -v github.com/marete/torrent@latest

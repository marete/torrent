torrent is a simple command-line BitTorrent client based on a
straight-forward usage of the excellent
[https://github.com/anacrolix/torrent](github.com/anacrolix/torrent)
library.

Usage:

    torrent -magnet '<magnet link>'
    torrent -magnet '<magnet link>' -dir <dest_dir>
    
By default torrent writes the downloaded file in the users `Downloads`
directory. This can be changed with the `-dir` command-line option.

torrent current supports only magnet links (and not, for example,
torrent files).

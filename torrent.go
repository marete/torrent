package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

// logger is for application logging in torrent.go. Configured at INFO level.
var logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// libLogger is for the upstream torrent library. Configured at ERROR level
// to suppress noisy INFO/WARN messages from the library.
var libLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	AddSource: true,
	Level:     slog.LevelError,
}))

var (
	dir = flag.String("dir", filepath.Join(mustGetHomeDir(), "Downloads"),
		"destination directory for torrent data")
	debug       = flag.Bool("debug", false, "enable debug logging")
	magnet      = flag.String("magnet", "", "magnet link")
	torrentFile = flag.String("file", "", "torrent file")
	dLimit      = flag.Float64("download_bandwidth_limit", 0, "limit total download bandwidth to this in Mega Bits/s. If zero or negative, impose no limit.")
	uLimit      = flag.Float64("upload_bandwidth_limit", 0, "limit total upload bandwidth to this in Mega Bits/s. If zero or negative, impose no limit.")
)

func limiter(l float64) *rate.Limiter {
	if l <= 0 {
		return nil
	}

	return rate.NewLimiter(rate.Limit((l/8)*1e6), 16<<10)
}

func mustGetHomeDir() string {
	u, err := user.Current()
	if err != nil {
		logger.Error("failed to get current user", "err", err)
		os.Exit(1)
	}

	return u.HomeDir
}

func abs(n int64) int64 {
	if n >= 0 {
		return n
	}

	return -n
}

func printBytes(n int64) string {
	switch {
	case abs(n) < 1e3:
		return fmt.Sprintf("%d B", n)
	case abs(n) < 1e6:
		return fmt.Sprintf("%.3f KB", float64(n)/1e3)
	case abs(n) < 1e9:
		return fmt.Sprintf("%.3f MB", float64(n)/1e6)
	case abs(n) < 1e12:
		return fmt.Sprintf("%.3f GB", float64(n)/1e9)
	default:
		return fmt.Sprintf("%.3f TB", float64(n)/1e12)
	}

}

func main() {
	flag.Parse()

	cfg := torrent.NewDefaultClientConfig()
	cfg.Slogger = libLogger
	cfg.DataDir = *dir
	cfg.Debug = *debug
	// Let the kernel provide the next free
	// port, which should allow multiple invocations of this tool to be
	// run concurrently without the user being forced to specify a port.
	cfg.ListenPort = 0
	client, err := torrent.NewClient(cfg)
	if err != nil {
		logger.Error("failed to create torrent client", "err", err)
		os.Exit(1)
	}
	defer client.Close()

	// Bandwidth limits
	dl := limiter(*dLimit)
	if dl != nil {
		cfg.DownloadRateLimiter = dl
	}
	ul := limiter(*uLimit)
	if ul != nil {
		cfg.UploadRateLimiter = ul
	}

	t, err := client.AddMagnet(*magnet)
	if err != nil {
		magErr := err
		t, err = client.AddTorrentFromFile(*torrentFile)
		if err != nil {
			logger.Error("a valid magnet link or torrent file must be provided",
				"magnetErr", magErr,
				"torrentFileErr", err)
			os.Exit(1)
		}
	}

	terminateReqCh := make(chan os.Signal, 1)
	signal.Notify(terminateReqCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	ignoredSignalsCh := make(chan os.Signal, 1)
	signal.Notify(ignoredSignalsCh, syscall.SIGHUP)

	select {
	case <-t.GotInfo():
	case <-terminateReqCh:
		client.Close()
		os.Exit(1)
	}

	t.DownloadAll()

	doneCh := make(chan bool, 1)
	go func() {
		doneCh <- client.WaitAll()
	}()

	for {
		select {
		case status := <-doneCh:
			if status {
				logger.Info("all torrents downloaded")
			} else {
				logger.Info("torrent download interrupted")
				client.Close()
				os.Exit(1)
			}
			return
		case <-time.After(time.Second * 3):
			done := t.BytesCompleted()
			miss := t.BytesMissing()
			logger.Info("download progress",
				"completed", printBytes(done),
				"total", printBytes(done+miss),
				"percent", fmt.Sprintf("%.2f%%", 100*float64(done)/float64(done+miss)))
		case <-terminateReqCh:
			logger.Info("closing all clients on termination signal")
			client.Close()
			logger.Info("terminating")
			os.Exit(1)
		case <-ignoredSignalsCh:
			continue
		}
	}
}

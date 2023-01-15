package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/golang/glog"
	"golang.org/x/time/rate"
)

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

	return rate.NewLimiter(rate.Limit((l/8)*1e6), 32<<10)
}

func mustGetHomeDir() string {
	u, err := user.Current()
	if err != nil {
		glog.Fatalf("user.Current(): %v", err)
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
	defer glog.Flush()

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = *dir
	cfg.Debug = *debug
	// Let the kernel provide the next free
	// port, which should allow multiple invocations of this tool to be
	// run concurrently without the user being forced to specify a port.
	cfg.ListenPort = 0
	client, err := torrent.NewClient(cfg)
	if err != nil {
		glog.Exitf("torrent.NewClient(): %v", err)
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
			glog.Exitf("A valid magnet link OR torrent file must be provided. client.AddMagnet() %v, client.AddTorrentFromFile() %v", magErr, err)
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
				glog.Infof("All torrents downloaded")
			} else {
				glog.Infof("Torrent download interrupted")
				client.Close()
				os.Exit(1)
			}
			return
		case <-time.After(time.Second * 3):
			done := t.BytesCompleted()
			miss := t.BytesMissing()
			log.Printf("%s / %s downloaded (%5.2f%%)",
				printBytes(done),
				printBytes(done+miss),
				100*float64(done)/(float64(done+miss)))
		case <-terminateReqCh:
			glog.Infof("Closing all clients on termination signal")
			client.Close()
			glog.Infof("Terminating")
			os.Exit(1)
		case <-ignoredSignalsCh:
			continue
		}
	}
}

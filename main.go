package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anacrolix/torrent"
	humanize "github.com/dustin/go-humanize"
	"github.com/golang/glog"
)

var (
	dir = flag.String("dir", filepath.Join(mustGetHomeDir(), "Downloads"),
		"destination directory for torrent data")
	debug  = flag.Bool("debug", false, "enable debug logging")
	magnet = flag.String("magnet", "", "magnet link")
)

func mustGetHomeDir() string {
	u, err := user.Current()
	if err != nil {
		glog.Fatalf("user.Current(): %v", err)
	}

	return u.HomeDir
}

func main() {
	flag.Parse()
	defer glog.Flush()

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = *dir
	cfg.Debug = *debug

	client, err := torrent.NewClient(cfg)
	if err != nil {
		glog.Exitf("torrent.NewClient(): %v", err)
	}
	defer client.Close()

	t, err := client.AddMagnet(*magnet)
	if err != nil {
		glog.Exitf("client.AddMagnet(): %v", err)
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
				humanize.Bytes(uint64(done)),
				humanize.Bytes(uint64(done+miss)),
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

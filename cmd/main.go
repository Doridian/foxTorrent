package main

import (
	"io"
	"log"
	"net/url"
	"os"

	"github.com/Doridian/foxTorrent/sideband/metainfo"
	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
	"github.com/Doridian/foxTorrent/sideband/tracker/httpproto"
	"github.com/Doridian/foxTorrent/sideband/tracker/udpproto"
)

func announceSupported(parsedUrl *url.URL) bool {
	if parsedUrl.Hostname() == "tracker.coppersurfer.tk" {
		return false
	}
	return parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" || parsedUrl.Scheme == "udp"
}

func main() {
	fh, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	data, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	meta, err := metainfo.Decode(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", meta)

	var totalLen uint64 = 0
	for _, file := range meta.Info.Files {
		totalLen += file.Length
	}

	log.Printf("totalLen: %d", totalLen)
	if totalLen == 0 {
		panic("len 0")
	}

	state := &announce.TorrentState{
		PeerID:     "foxTorrent dummyPeer",
		Port:       1337,
		Uploaded:   0,
		Downloaded: 0,
		Left:       totalLen,
		Meta:       meta,
	}

	var announceUrl *url.URL
	for _, announceList := range meta.AnnounceList {
		for _, announce := range announceList {
			parsedUrl, err := url.Parse(announce)
			if err != nil {
				continue
			}
			if announceSupported(parsedUrl) {
				announceUrl = parsedUrl
				break
			}
		}
		if announceUrl != nil {
			break
		}
	}

	var announcer announce.Announcer

	switch announceUrl.Scheme {
	case "http", "https":
		announcer, err = httpproto.NewClient(*announceUrl)
		if err != nil {
			log.Fatal(err)
		}
	case "udp":
		announcer, err = udpproto.NewClient(announceUrl.Host)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unsupported scheme: %s", announceUrl.Scheme)
	}

	log.Printf("Connecting to announce at %v", announceUrl)

	err = announcer.Connect()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected. Announcing event started")

	res, err := announcer.AnnounceEvent(state, announce.EventStarted)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", res)
}

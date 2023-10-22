package main

import (
	"io"
	"log"
	"net/url"
	"os"

	"github.com/Doridian/foxTorrent/pkg/metainfo"
	"github.com/Doridian/foxTorrent/pkg/torrent/state"
	"github.com/Doridian/foxTorrent/pkg/tracker"
	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
	"github.com/Workiva/go-datastructures/bitarray"
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

	totalLen := meta.TotalLength()
	log.Printf("totalLen: %d", totalLen)
	if totalLen == 0 {
		panic("totalLen 0")
	}

	state := &state.State{
		PeerID:     "foxTorrent dummyPeer",
		Port:       1337,
		Uploaded:   0,
		Downloaded: 0,
		Left:       totalLen,
		InfoHash:   meta.InfoHash,
		Pieces:     bitarray.NewBitArray(uint64(len(meta.Info.Pieces))),
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

	announcer, err := tracker.CreateFromURL(*announceUrl)
	if err != nil {
		log.Fatal(err)
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

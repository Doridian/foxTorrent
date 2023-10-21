package main

import (
	"io"
	"log"
	"os"

	"github.com/Doridian/foxTorrent/sideband/announce"
	"github.com/Doridian/foxTorrent/sideband/announce/httpproto"
	"github.com/Doridian/foxTorrent/sideband/metainfo"
)

func announceSupported(urlStr string) bool {
	return urlStr[:5] == "http:" || urlStr[:6] == "https:"
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

	var totalLen int64 = 0
	for _, file := range meta.Info.Files {
		totalLen += file.Length
	}

	log.Printf("totalLen: %d", totalLen)
	if totalLen == 0 {
		panic("len 0")
	}

	info := &announce.ClientInfo{
		PeerID:     "foxTorrent dummyPeer",
		Port:       1337,
		Uploaded:   0,
		Downloaded: 0,
		Left:       totalLen,
	}

	var announceUrl string
	for _, announceList := range meta.AnnounceList {
		for _, announce := range announceList {
			if announceSupported(announce) {
				announceUrl = announce
				break
			}
		}
		if announceUrl != "" {
			break
		}
	}

	res, err := httpproto.SendAnnounceEvent(announceUrl, "started", info, meta)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", res)
}

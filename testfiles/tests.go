package testfiles

import (
	_ "embed"
)

//go:embed ubuntu-23.10-live-server-amd64.iso.torrent
var Ubuntu2310LiveServerAMD64IsoTorrent []byte

//go:embed ubuntu-23.10-live-server-amd64.iso.announce
var Ubuntu2310LiveServerAMD64IsoAnnounce []byte

//go:embed ubuntu-23.10-live-server-amd64.iso.announce.ipv4
var Ubuntu2310LiveServerAMD64IsoAnnounceIpv4 []byte

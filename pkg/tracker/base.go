package tracker

import (
	"github.com/Doridian/foxTorrent/pkg/metainfo"
	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
)

type Announcer interface {
	Announce(state *TorrentState) (*announce.Announce, error)
	AnnounceEvent(state *TorrentState, event uint32) (*announce.Announce, error)
	Connect() error
}

type TorrentState struct {
	PeerID string
	Port   uint16

	Uploaded   uint64
	Downloaded uint64
	Left       uint64

	Meta *metainfo.Metainfo
}

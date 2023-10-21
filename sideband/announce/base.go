package announce

import (
	"net"

	"github.com/Doridian/foxTorrent/sideband/metainfo"
)

type Peer struct {
	IP     net.IP
	Port   uint16
	PeerID string
}

type Announce struct {
	WarningMessage string

	Interval    uint32
	MinInterval uint32

	TrackerID string

	Complete   uint32
	Incomplete uint32

	Peers []Peer
}

const (
	EventNone      = 0
	EventCompleted = 1
	EventStarted   = 2
	EventStopped   = 3
)

type ClientInfo struct {
	PeerID string
	Port   uint16

	Uploaded   uint64
	Downloaded uint64
	Left       uint64

	Meta *metainfo.Metainfo
}

type Announcer interface {
	Announce(info *ClientInfo) (*Announce, error)
	AnnounceEvent(info *ClientInfo, event uint32) (*Announce, error)
	Connect() error
}

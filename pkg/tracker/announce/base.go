package announce

import (
	"net"

	"github.com/Doridian/foxTorrent/pkg/torrent"
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

type Announcer interface {
	Announce(state *torrent.State) (*Announce, error)
	AnnounceEvent(state *torrent.State, event uint32) (*Announce, error)
	Connect() error
}

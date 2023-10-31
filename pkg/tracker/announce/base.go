package announce

import (
	"net"
	"time"

	"github.com/Doridian/foxTorrent/pkg/torrent/state"
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

	NextAnnounce    time.Time
	NextMinAnnounce time.Time

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
	Announce(state *state.State) (*Announce, error)
	AnnounceEvent(state *state.State, event uint32) (*Announce, error)
	Connect() error
}

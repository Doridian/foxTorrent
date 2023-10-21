package announce

import "net"

type Peer struct {
	IP     net.IP
	Port   int64
	PeerID string
}

type Announce struct {
	WarningMessage string

	Interval    int64
	MinInterval int64

	TrackerID string

	Complete   int64
	Incomplete int64

	Peers []Peer
}

type ClientInfo struct {
	PeerID     string
	TrackerID  string
	Port       int64
	Uploaded   int64
	Downloaded int64
	Left       int64
}

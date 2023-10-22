package state

type State struct {
	PeerID string
	Port   uint16

	Uploaded   uint64
	Downloaded uint64
	Left       uint64

	InfoHash []byte
}

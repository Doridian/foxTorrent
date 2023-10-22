package state

import "github.com/Workiva/go-datastructures/bitarray"

type State struct {
	PeerID string
	Port   uint16

	Uploaded   uint64
	Downloaded uint64
	Left       uint64

	InfoHash []byte
	Pieces   bitarray.BitArray
}

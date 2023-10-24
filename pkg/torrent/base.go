package torrent

import (
	"bytes"
	"net"
	"sync"

	"github.com/Workiva/go-datastructures/bitarray"
)

type SendPieceReply = func([]byte) error
type OnPieceRequestHandler = func(index uint32, begin uint32, length uint32, reply SendPieceReply) error
type OnPieceCancelHandler = func(index uint32, begin uint32, length uint32) error

type Connection struct {
	conn net.Conn

	remotePeerID string
	localPeerID  string
	infoHash     []byte

	infoHashValidator InfoHashValidator

	localChoking     bool
	remoteChoking    bool
	localInterested  bool
	remoteInterested bool

	localHave       bitarray.BitArray
	remoteHave      bitarray.BitArray
	canSendBitfield bool

	pieceRequests    map[uint64]*PieceRequest
	pieceRequestLock sync.Mutex

	OnPieceRequest OnPieceRequestHandler
	OnPieceCancel  OnPieceCancelHandler
}

type InfoHashValidator func(infoHash []byte) (bool, error)

func ServeAsInitiator(conn net.Conn, infoHash []byte, localPeerID string, remotePeerID string) (*Connection, error) {
	btConn := &Connection{
		conn: conn,

		localPeerID:  localPeerID,
		remotePeerID: remotePeerID,
		infoHash:     infoHash,

		localChoking:     true,
		remoteChoking:    true,
		localInterested:  false,
		remoteInterested: false,

		localHave:       bitarray.NewBitArray(0),
		remoteHave:      bitarray.NewBitArray(0),
		canSendBitfield: true,

		pieceRequests: make(map[uint64]*PieceRequest),
	}
	btConn.infoHashValidator = btConn.infoHashValidatorSelf

	err := btConn.TransmitHandshake()
	if err != nil {
		return nil, err
	}
	err = btConn.ReceiveHandshake(false)
	if err != nil {
		return nil, err
	}

	return btConn, nil
}

func ServeAsRecipient(conn net.Conn, infoHashValidator InfoHashValidator, localPeerID string, remotePeerID string) (*Connection, error) {
	btConn := &Connection{
		conn: conn,

		localPeerID:  localPeerID,
		remotePeerID: remotePeerID,
		infoHash:     nil,

		infoHashValidator: infoHashValidator,

		localChoking:     true,
		remoteChoking:    true,
		localInterested:  false,
		remoteInterested: false,

		localHave:       bitarray.NewBitArray(0),
		remoteHave:      bitarray.NewBitArray(0),
		canSendBitfield: true,

		pieceRequests: make(map[uint64]*PieceRequest),
	}

	err := btConn.ReceiveHandshake(true)
	if err != nil {
		return nil, err
	}

	return btConn, nil
}

func (c *Connection) infoHashValidatorSelf(infoHash []byte) (bool, error) {
	return bytes.Equal(infoHash, c.infoHash), nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) RemotePeerID() string {
	return c.remotePeerID
}

func (c *Connection) LocalPeerID() string {
	return c.localPeerID
}

func (c *Connection) InfoHash() []byte {
	return c.infoHash
}

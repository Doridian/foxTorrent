package torrent

import (
	"bytes"
	"net"
	"sync"

	"github.com/Doridian/foxTorrent/pkg/bitfield"
)

type OnPieceRequestHandler func(conn *Connection, index uint32, begin uint32, length uint32, reply SendPieceReply) error
type OnPieceCancelHandler func(conn *Connection, index uint32, begin uint32, length uint32)
type InfoHashValidatorHandler func(conn *Connection, infoHash []byte) (bool, error)

type Connection struct {
	conn net.Conn

	remotePeerID string
	localPeerID  string
	infoHash     []byte

	localChoking     bool
	remoteChoking    bool
	localInterested  bool
	remoteInterested bool

	localHave          *bitfield.Bitfield
	remoteHave         *bitfield.Bitfield
	canSendBitfield    bool
	canReceiveBitfield bool

	pieceRequests    map[uint64]*PieceRequest
	pieceRequestLock sync.Mutex

	OnPieceRequest     OnPieceRequestHandler
	OnPieceCancel      OnPieceCancelHandler
	OnRemoteChoke      func(conn *Connection, choking bool)
	OnRemoteInterested func(conn *Connection, interested bool)
	InfoHashValidator  InfoHashValidatorHandler
	OnHaveUpdated      func(conn *Connection, piece int64)
}

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

		localHave:          nil,
		remoteHave:         nil,
		canSendBitfield:    true,
		canReceiveBitfield: true,

		pieceRequests: make(map[uint64]*PieceRequest),
	}
	btConn.InfoHashValidator = btConn.infoHashValidatorSelf

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

func ServeAsRecipient(conn net.Conn, infoHashValidator InfoHashValidatorHandler, localPeerID string, remotePeerID string) (*Connection, error) {
	btConn := &Connection{
		conn: conn,

		localPeerID:  localPeerID,
		remotePeerID: remotePeerID,
		infoHash:     nil,

		InfoHashValidator: infoHashValidator,

		localChoking:     true,
		remoteChoking:    true,
		localInterested:  false,
		remoteInterested: false,

		localHave:          nil,
		remoteHave:         nil,
		canSendBitfield:    true,
		canReceiveBitfield: true,

		pieceRequests: make(map[uint64]*PieceRequest),
	}

	err := btConn.ReceiveHandshake(true)
	if err != nil {
		return nil, err
	}

	return btConn, nil
}

func (c *Connection) infoHashValidatorSelf(conn *Connection, infoHash []byte) (bool, error) {
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

func (c *Connection) SetPieceCount(pieces uint64) {
	c.remoteHave = bitfield.NewBitfield(pieces)
	c.localHave = bitfield.NewBitfield(pieces)
}

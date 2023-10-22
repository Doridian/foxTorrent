package torrent

import (
	"bytes"
	"net"
)

type State struct {
	PeerID string
	Port   uint16

	Uploaded   uint64
	Downloaded uint64
	Left       uint64

	InfoHash []byte
}

type Connection struct {
	conn net.Conn

	infoHashValidator InfoHashValidator

	remotePeerID string
	localPeerID  string
	infoHash     []byte
}

type InfoHashValidator func(infoHash []byte) (bool, error)

func ServeAsInitiator(conn net.Conn, state *State) (*Connection, error) {
	ourInfoHash := state.InfoHash
	btConn := &Connection{
		conn:     conn,
		infoHash: ourInfoHash,
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

func ServeAsRecipient(conn net.Conn, infoHashValidator InfoHashValidator) (*Connection, error) {
	btConn := &Connection{
		conn:              conn,
		infoHash:          nil,
		infoHashValidator: infoHashValidator,
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

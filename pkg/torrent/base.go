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

	handshake []byte
}

type Connection struct {
	conn         net.Conn
	stateLookup  StateLookup
	state        *State
	remotePeerID string
}

type StateLookup func(infoHash []byte) (*State, error)

func ServeAsInitiator(conn net.Conn, state *State) (*Connection, error) {
	btConn := &Connection{
		conn:        conn,
		stateLookup: nil,
		state:       state,
	}

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

func ServeAsRecipient(conn net.Conn, stateLookup StateLookup) (*Connection, error) {
	btConn := &Connection{
		conn:        conn,
		stateLookup: stateLookup,
		state:       nil,
	}

	err := btConn.ReceiveHandshake(true)
	if err != nil {
		return nil, err
	}

	return btConn, nil
}

func (c *Connection) GetState(infoHash []byte) (*State, error) {
	if c.state == nil {
		if c.stateLookup == nil {
			return nil, nil
		}
		state, err := c.stateLookup(infoHash)
		if err != nil {
			return nil, err
		}
		c.state = state
		return state, nil
	}

	if !bytes.Equal(c.state.InfoHash, infoHash) {
		return nil, nil
	}
	return c.state, nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

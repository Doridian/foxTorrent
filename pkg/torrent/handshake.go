package torrent

import (
	"errors"
	"io"
)

const ProtocolStr = "BitTorrent protocol"
const ProtocolStrLen = len(ProtocolStr)
const HandshakeLenBeforeInfoHash = 1 + ProtocolStrLen + 8
const HandshakeLenMin = HandshakeLenBeforeInfoHash + 20
const HandshakeLenFull = HandshakeLenMin + 20

var reservedBytes = []byte{0, 0, 0, 0, 0, 0, 0, 0}

var ErrInvalidHandshake = errors.New("invalid handshake")

func (c *Connection) TransmitHandshake() error {
	handshake := make([]byte, 0, 68)
	handshake = append(handshake, uint8(ProtocolStrLen))
	handshake = append(handshake, ProtocolStr...)
	handshake = append(handshake, reservedBytes...)
	handshake = append(handshake, c.infoHash...)
	handshake = append(handshake, c.localPeerID...)

	_, err := c.conn.Write(handshake)
	return err
}

func (c *Connection) ReceiveHandshake(respondAfterInfoHash bool) error {
	buf := make([]byte, HandshakeLenFull)
	bufPos, err := c.conn.Read(buf)
	if err != nil {
		return err
	}

	readProtocolLen := int(buf[0])
	if readProtocolLen != ProtocolStrLen {
		return ErrInvalidHandshake
	}

	if bufPos < HandshakeLenMin {
		readLen, err := io.ReadAtLeast(c.conn, buf[bufPos:], HandshakeLenMin-bufPos)
		if err != nil {
			return err
		}
		bufPos += readLen
	}

	readInfoHash := buf[HandshakeLenBeforeInfoHash:HandshakeLenMin]

	infoHashValid, err := c.InfoHashValidator(c, readInfoHash)
	if err != nil {
		return err
	}

	if !infoHashValid {
		return ErrInvalidHandshake
	}

	c.infoHash = readInfoHash
	c.InfoHashValidator = c.infoHashValidatorSelf

	if respondAfterInfoHash {
		err = c.TransmitHandshake()
		if err != nil {
			return err
		}
	}

	_, err = io.ReadFull(c.conn, buf[bufPos:HandshakeLenFull])
	if err != nil {
		return err
	}

	readPeerID := string(buf[HandshakeLenMin:HandshakeLenFull])
	if c.remotePeerID != "" && c.remotePeerID != readPeerID {
		return ErrInvalidHandshake
	}
	c.remotePeerID = readPeerID

	return nil
}

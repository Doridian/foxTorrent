package torrent

import (
	"encoding/binary"

	"github.com/Doridian/foxTorrent/pkg/torrent/state"
)

const BitBufferLen = 512

func (c *Connection) SendHaveState(curState *state.State) error {
	piecesSent := curState.Pieces
	if c.canSendBitfield {
		err := c.WritePacket(&Packet{
			ID:      PacketBitfield,
			Payload: piecesSent.GetData(),
		})
		if err != nil {
			return err
		}

		c.localHave = piecesSent
		return nil
	}

	pieceDelta := piecesSent
	if c.localHave != nil {
		pieceDelta = piecesSent.Delta(c.localHave)
	}

	setBitBuffer := make([]uint64, 0, BitBufferLen)
	var pos uint64 = 0
	for {
		setBitBuffer = setBitBuffer[:0]
		pieceDelta.GetSetBits(pos, setBitBuffer)

		for _, piece := range setBitBuffer {
			err := c.SendHavePiece(uint32(piece))
			if err != nil {
				return err
			}
		}

		if len(setBitBuffer) < BitBufferLen {
			break
		}
		pos += BitBufferLen
	}

	c.localHave = piecesSent
	return nil
}

func (c *Connection) SendHavePiece(piece uint32) error {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, piece)

	return c.WritePacket(&Packet{
		ID:      PacketHave,
		Payload: payload,
	})
}

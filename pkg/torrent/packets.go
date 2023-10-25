package torrent

import (
	"encoding/binary"
	"errors"
	"io"
	"log"

	"github.com/Doridian/foxTorrent/pkg/bitfield"
)

type SendPieceReply func(piece []byte) error

type Packet struct {
	ID      uint8
	Payload []byte
}

const (
	PacketChoke         = 0
	PacketUnchoke       = 1
	PacketInterested    = 2
	PacketNotInterested = 3
	PacketHave          = 4
	PacketBitfield      = 5
	PacketRequest       = 6
	PacketPiece         = 7
	PacketCancel        = 8
	PacketPort          = 9
)

func (c *Connection) ReadPacket() (*Packet, error) {
	packetLen := make([]byte, 4)
	_, err := io.ReadFull(c.conn, packetLen)
	if err != nil {
		return nil, err
	}

	packetLenInt := binary.BigEndian.Uint32(packetLen)
	if packetLenInt == 0 { // keep-alive packet
		return nil, nil
	}

	packet := make([]byte, packetLenInt)
	_, err = io.ReadFull(c.conn, packet)
	if err != nil {
		return nil, err
	}

	c.canReceiveBitfield = false

	return &Packet{
		ID:      packet[0],
		Payload: packet[1:],
	}, nil
}

func (c *Connection) WritePacket(packet *Packet) error {
	c.canSendBitfield = false

	payload := make([]byte, 0, 4+1+len(packet.Payload))
	payload = binary.BigEndian.AppendUint32(payload, uint32(1+len(packet.Payload)))
	payload = append(payload, packet.ID)
	payload = append(payload, packet.Payload...)
	_, err := c.conn.Write(payload)
	return err
}

func (c *Connection) Serve() error {
	defer c.Close()
	return c.serve()
}

func (c *Connection) serve() error {
	for {
		canReceiveBitfield := c.canReceiveBitfield
		packet, err := c.ReadPacket()
		if err != nil {
			return err
		}

		if packet == nil {
			continue
		}

		switch packet.ID {
		case PacketChoke:
			c.remoteChoking = true
			if c.OnRemoteChoke != nil {
				go c.OnRemoteChoke(c, true)
			}

		case PacketUnchoke:
			c.remoteChoking = false
			if c.OnRemoteChoke != nil {
				go c.OnRemoteChoke(c, false)
			}

		case PacketInterested:
			c.remoteInterested = true
			if c.OnRemoteInterested != nil {
				go c.OnRemoteInterested(c, true)
			}

		case PacketNotInterested:
			c.remoteInterested = false
			if c.OnRemoteInterested != nil {
				go c.OnRemoteInterested(c, false)
			}

		case PacketHave:
			piece := binary.BigEndian.Uint32(packet.Payload)

			if c.remoteHave != nil {
				c.remoteHave.SetBit(uint64(piece))

				if c.OnHaveUpdated != nil {
					go c.OnHaveUpdated(c, int64(piece))
				}
			}

		case PacketBitfield:
			if !canReceiveBitfield {
				return errors.New("unexpected bitfield packet")
			}

			if c.remoteHave != nil {
				if len(packet.Payload) != len(c.remoteHave.GetData()) {
					return errors.New("invalid bitfield length")
				}

				c.remoteHave = bitfield.NewBitfieldFromBytes(packet.Payload)

				if c.OnHaveUpdated != nil {
					go c.OnHaveUpdated(c, -1)
				}
			}

		case PacketRequest:
			index := binary.BigEndian.Uint32(packet.Payload[:4])
			begin := binary.BigEndian.Uint32(packet.Payload[4:8])
			length := binary.BigEndian.Uint32(packet.Payload[8:12])

			if c.localChoking {
				return errors.New("got request while choked")
			}

			if !c.remoteInterested {
				return errors.New("got request while not interested")
			}

			if c.OnPieceRequest == nil {
				log.Printf("got request for piece %d, but no handler is registered", index)
			} else {
				go func(index uint32, begin uint32, length uint32) {
					err := c.OnPieceRequest(c, index, begin, length, func(data []byte) error {
						payload := make([]byte, 0, 8+len(data))
						payload = binary.BigEndian.AppendUint32(payload, index)
						payload = binary.BigEndian.AppendUint32(payload, begin)
						payload = append(payload, data...)
						return c.WritePacket(&Packet{
							ID:      PacketPiece,
							Payload: payload,
						})
					})
					if err != nil {
						log.Printf("error handling piece request: %v", err)
					}
				}(index, begin, length)
			}

		case PacketPiece:
			index := binary.BigEndian.Uint32(packet.Payload[:4])
			begin := binary.BigEndian.Uint32(packet.Payload[4:8])
			block := packet.Payload[8:]

			go func(index uint32, begin uint32, block []byte) {
				err := c.onPieceData(index, begin, block)
				if err != nil {
					log.Printf("error handling piece data: %v", err)
				}
			}(index, begin, block)

		case PacketCancel:
			index := binary.BigEndian.Uint32(packet.Payload[:4])
			begin := binary.BigEndian.Uint32(packet.Payload[4:8])
			length := binary.BigEndian.Uint32(packet.Payload[8:12])

			if c.OnPieceCancel == nil {
				log.Printf("got cancel for piece %d, but no handler is registered", index)
			} else {
				err := c.OnPieceCancel(c, index, begin, length)
				if err != nil {
					return err
				}
			}
		}
	}
}

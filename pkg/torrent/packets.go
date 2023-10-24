package torrent

import (
	"encoding/binary"
	"errors"
	"io"
	"log"

	"github.com/Workiva/go-datastructures/bitarray"
)

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
	for {
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
		case PacketUnchoke:
			c.remoteChoking = false
			go c.requestNextPiece()
		case PacketInterested:
			c.remoteInterested = true
		case PacketNotInterested:
			c.remoteInterested = false
		case PacketHave:
			piece := binary.BigEndian.Uint32(packet.Payload)
			c.remoteHave.SetBit(uint64(piece))
		case PacketBitfield:
			if !c.remoteHave.IsEmpty() {
				return errors.New("unexpected bitfield packet")
			}

			newRemoteHave := bitarray.NewBitArray(uint64(len(packet.Payload)) * 8)

			for i := 0; i < len(packet.Payload); i++ {
				for j := 0; j < 8; j++ {
					if packet.Payload[i]&(1<<uint(7-j)) != 0 {
						newRemoteHave.SetBit(uint64(i*8 + j))
					}
				}
			}

			c.remoteHave = newRemoteHave
		case PacketRequest:
			index := binary.BigEndian.Uint32(packet.Payload[:4])
			begin := binary.BigEndian.Uint32(packet.Payload[4:8])
			length := binary.BigEndian.Uint32(packet.Payload[8:12])

			log.Printf("Received request for index %d, begin %d, length %d", index, begin, length)
		case PacketPiece:
			index := binary.BigEndian.Uint32(packet.Payload[:4])
			begin := binary.BigEndian.Uint32(packet.Payload[4:8])
			block := packet.Payload[8:]

			log.Printf("Received piece for index %d, begin %d, block length %d", index, begin, len(block))
			err := c.onPieceData(index, begin, block)
			if err != nil {
				return err
			}

		case PacketCancel:
			index := binary.BigEndian.Uint32(packet.Payload[:4])
			begin := binary.BigEndian.Uint32(packet.Payload[4:8])
			length := binary.BigEndian.Uint32(packet.Payload[8:12])

			log.Printf("Received cancel for index %d, begin %d, length %d", index, begin, length)
		}
	}
}

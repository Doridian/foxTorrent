package torrent

import (
	"encoding/binary"
	"io"
)

type Packet struct {
	ID      uint8
	Payload []byte
}

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
	payload := make([]byte, 0, 4+1+len(packet.Payload))
	payload = binary.BigEndian.AppendUint32(payload, uint32(1+len(packet.Payload)))
	payload = append(payload, packet.ID)
	payload = append(payload, packet.Payload...)
	_, err := c.conn.Write(payload)
	return err
}

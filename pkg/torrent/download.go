package torrent

import (
	"encoding/binary"
	"log"
)

type PieceRequest struct {
	Index  uint32
	Begin  uint32
	Length uint32

	Callback func([]byte)
}

func mapIndex(index uint32, begin uint32) uint64 {
	return uint64(index)<<32 | uint64(begin)
}

func (c *Connection) RequestPiece(request *PieceRequest) error {
	c.pieceRequestLock.Lock()
	defer c.pieceRequestLock.Unlock()
	c.pieceRequests[mapIndex(request.Index, request.Begin)] = request

	c.setInterested(true)

	payload := make([]byte, 0, 12)
	payload = binary.BigEndian.AppendUint32(payload, request.Index)
	payload = binary.BigEndian.AppendUint32(payload, request.Begin)
	payload = binary.BigEndian.AppendUint32(payload, request.Length)
	return c.WritePacket(&Packet{
		ID:      PacketRequest,
		Payload: payload,
	})
}

func (c *Connection) CancelPiece(request *PieceRequest) error {
	c.pieceRequestLock.Lock()
	defer c.pieceRequestLock.Unlock()

	delete(c.pieceRequests, mapIndex(request.Index, request.Begin))

	payload := make([]byte, 0, 12)
	payload = binary.BigEndian.AppendUint32(payload, request.Index)
	payload = binary.BigEndian.AppendUint32(payload, request.Begin)
	payload = binary.BigEndian.AppendUint32(payload, request.Length)
	return c.WritePacket(&Packet{
		ID:      PacketCancel,
		Payload: payload,
	})
}

func (c *Connection) GetPieceQueueLength() int {
	return len(c.pieceRequests)
}

func (c *Connection) onPieceData(index uint32, begin uint32, data []byte) error {
	c.pieceRequestLock.Lock()
	defer c.pieceRequestLock.Unlock()

	pieceMapIndex := mapIndex(index, begin)
	pieceRequest := c.pieceRequests[pieceMapIndex]
	if pieceRequest == nil {
		return nil
	}

	go pieceRequest.Callback(data)

	delete(c.pieceRequests, pieceMapIndex)

	if c.GetPieceQueueLength() == 0 {
		go func() {
			err := c.setInterested(false)
			if err != nil {
				log.Printf("error setting interested to false: %v", err)
			}
		}()
	}

	return nil
}

package torrent

import (
	"encoding/binary"
)

type PieceRequest struct {
	Index  uint32
	Begin  uint32
	Length uint32

	Callback func([]byte)
}

func (c *Connection) RequestPiece(request *PieceRequest) error {
	c.pieceRequestLock.Lock()
	defer c.pieceRequestLock.Unlock()

	indexQueue := c.pieceRequests[request.Index]
	if indexQueue == nil {
		indexQueue = make(map[uint32]*PieceRequest)
		c.pieceRequests[request.Index] = indexQueue
	}
	indexQueue[request.Begin] = request

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

	indexQueue := c.pieceRequests[request.Index]
	if indexQueue != nil {
		delete(indexQueue, request.Begin)
		if len(indexQueue) == 0 {
			delete(c.pieceRequests, request.Index)
		}
	}

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

	indexQueue := c.pieceRequests[index]
	if indexQueue == nil {
		return nil
	}

	handledPieceRequest := indexQueue[begin]
	if handledPieceRequest == nil {
		return nil
	}

	go handledPieceRequest.Callback(data)

	delete(indexQueue, begin)
	if len(indexQueue) == 0 {
		delete(c.pieceRequests, index)
	}

	return nil
}

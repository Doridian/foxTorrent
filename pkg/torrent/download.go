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

// TODO: Important! We need multiple concurrent requests in-flight at once.

func (c *Connection) RequestPiece(request *PieceRequest) error {
	c.pieceQueueLock.Lock()
	c.pieceRequestQueue = append(c.pieceRequestQueue, request)
	c.setInterested(true)
	c.pieceQueueLock.Unlock()

	return c.requestNextPiece()
}

func (c *Connection) CancelPiece(request *PieceRequest) error {
	c.pieceQueueLock.Lock()
	defer c.pieceQueueLock.Unlock()

	if request == c.currentPieceRequest {
		c.currentPieceRequest = nil
		c.pieceQueueLock.Unlock()

		payload := make([]byte, 0, 12)
		payload = binary.BigEndian.AppendUint32(payload, request.Index)
		payload = binary.BigEndian.AppendUint32(payload, request.Begin)
		payload = binary.BigEndian.AppendUint32(payload, request.Length)
		return c.WritePacket(&Packet{
			ID:      PacketCancel,
			Payload: payload,
		})
	}

	for i, queuedRequest := range c.pieceRequestQueue {
		if queuedRequest == request {
			c.pieceRequestQueue = append(c.pieceRequestQueue[:i], c.pieceRequestQueue[i+1:]...)
			break
		}
	}

	return nil
}

func (c *Connection) GetPieceQueueLength() int {
	return len(c.pieceRequestQueue)
}

func (c *Connection) requestNextPiece() error {
	c.pieceQueueLock.Lock()
	defer c.pieceQueueLock.Unlock()

	if len(c.pieceRequestQueue) == 0 {
		c.setInterested(false)
		return nil
	}

	if c.currentPieceRequest != nil {
		return nil
	}
	if c.remoteChoking {
		return nil
	}

	request := c.pieceRequestQueue[0]
	c.pieceRequestQueue = c.pieceRequestQueue[1:]
	c.currentPieceRequest = request

	payload := make([]byte, 0, 12)
	payload = binary.BigEndian.AppendUint32(payload, request.Index)
	payload = binary.BigEndian.AppendUint32(payload, request.Begin)
	payload = binary.BigEndian.AppendUint32(payload, request.Length)
	return c.WritePacket(&Packet{
		ID:      PacketRequest,
		Payload: payload,
	})
}

func (c *Connection) onPieceData(index uint32, begin uint32, data []byte) error {
	c.pieceQueueLock.Lock()
	defer c.pieceQueueLock.Unlock()

	handledPieceRequest := c.currentPieceRequest
	if begin != handledPieceRequest.Begin {
		return nil
	}
	if index != handledPieceRequest.Index {
		return nil
	}
	if len(data) != int(handledPieceRequest.Length) {
		return nil
	}
	c.currentPieceRequest = nil
	go handledPieceRequest.Callback(data)
	return nil
}

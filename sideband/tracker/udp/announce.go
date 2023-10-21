package udp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/Doridian/foxTorrent/sideband/tracker"
	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
)

func (c *UDPClient) Announce(state *tracker.TorrentState) (*announce.Announce, error) {
	return c.AnnounceEvent(state, announce.EventNone)
}

func (c *UDPClient) AnnounceEvent(state *tracker.TorrentState, event uint32) (*announce.Announce, error) {
	if c.conn == nil {
		return nil, errors.New("not connected")
	}

	announcePayload := make([]byte, 0, 82)
	announcePayload = append(announcePayload, state.Meta.InfoHash...)
	announcePayload = append(announcePayload, []byte(state.PeerID)...)
	announcePayload = binary.BigEndian.AppendUint64(announcePayload, state.Downloaded)
	announcePayload = binary.BigEndian.AppendUint64(announcePayload, state.Left)
	announcePayload = binary.BigEndian.AppendUint64(announcePayload, state.Uploaded)
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, event)
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, 0)  // IP
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, 0)  // "key"
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, 50) // numwant
	announcePayload = binary.BigEndian.AppendUint16(announcePayload, state.Port)

	announceResp, err := c.sendRecv(ActionAnnounce, announcePayload)
	if err != nil {
		return nil, err
	}

	if len(announceResp) < 12 {
		return nil, errors.New("short announce response")
	}

	peerLen := len(announceResp) - 12
	if peerLen%6 != 0 {
		return nil, fmt.Errorf("invalid peer length: %d", peerLen)
	}
	peerCount := peerLen / 6

	result := &announce.Announce{
		Interval:   binary.BigEndian.Uint32(announceResp[0:4]),
		Incomplete: binary.BigEndian.Uint32(announceResp[4:8]),
		Complete:   binary.BigEndian.Uint32(announceResp[8:12]),
		Peers:      make([]announce.Peer, 0, peerCount),
	}

	for i := 0; i < peerLen; i += 6 {
		result.Peers = append(result.Peers, announce.Peer{
			IP:   net.IP(announceResp[12+i : 16+i]),
			Port: binary.BigEndian.Uint16(announceResp[16+i : 18+i]),
		})
	}

	return result, nil
}

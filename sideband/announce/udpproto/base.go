package udpproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Doridian/foxTorrent/sideband/announce"
)

const (
	actionConnect  = 0
	actionAnnounce = 1
	actionScrape   = 2
	actionError    = 3
)

type UDPClient struct {
	addr string

	conn    *net.UDPConn
	udpAddr *net.UDPAddr

	connectionID uint64
}

func NewClient(addr string) (announce.Announcer, error) {
	return &UDPClient{
		addr: addr,
	}, nil
}

func (c *UDPClient) sendRecv(action uint32, payload []byte) ([]byte, error) {
	buffer := make([]byte, 65536)
	var err error

	transactionID := uint32(time.Now().UnixNano())

	packet := make([]byte, 0, 16+len(payload))
	packet = binary.BigEndian.AppendUint64(packet, c.connectionID)
	packet = binary.BigEndian.AppendUint32(packet, action)
	packet = binary.BigEndian.AppendUint32(packet, transactionID)
	packet = append(packet, payload...)

	for retries := 0; retries < 4; retries++ {
		_, err = c.conn.WriteToUDP(packet, c.udpAddr)
		if err != nil {
			return nil, err
		}

		err = c.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		if err != nil {
			return nil, err
		}

		for { // loop reading packets, ignoring ones we don't care about
			var readLen int
			var readAddr *net.UDPAddr
			readLen, readAddr, err = c.conn.ReadFromUDP(buffer)
			if err != nil {
				break
			}
			if !readAddr.IP.Equal(c.udpAddr.IP) || readAddr.Port != c.udpAddr.Port {
				continue
			}

			if readLen < 16 {
				continue
			}

			recvTransactionID := binary.BigEndian.Uint32(buffer[4:8])
			if recvTransactionID != transactionID {
				continue
			}

			recvAction := binary.BigEndian.Uint32(buffer[0:4])
			recvPayload := buffer[8:readLen]

			if recvAction == actionError { // error
				return nil, fmt.Errorf("tracker error: %s", recvPayload)
			}

			if recvAction != action {
				continue
			}

			return recvPayload, nil
		}
	}

	return nil, err
}

func (c *UDPClient) Connect() error {
	var err error
	c.udpAddr, err = net.ResolveUDPAddr("udp", c.addr)
	if err != nil {
		return err
	}
	c.conn, err = net.ListenUDP("udp", nil)
	if err != nil {
		return err
	}
	c.connectionID = 0x41727101980

	connectResp, err := c.sendRecv(actionConnect, []byte{})
	if err != nil {
		return err
	}

	if len(connectResp) < 8 {
		return errors.New("short connect response")
	}

	c.connectionID = binary.BigEndian.Uint64(connectResp[0:8])
	return nil
}

func (c *UDPClient) Announce(state *announce.TorrentState) (*announce.Announce, error) {
	return c.AnnounceEvent(state, announce.EventNone)
}

func (c *UDPClient) AnnounceEvent(state *announce.TorrentState, event uint32) (*announce.Announce, error) {
	if c.conn == nil {
		return nil, errors.New("not connected")
	}

	announcePayload := make([]byte, 0, 82)
	announcePayload = append(announcePayload, state.Meta.InfoHash[:]...)
	announcePayload = append(announcePayload, []byte(state.PeerID)...)
	announcePayload = binary.BigEndian.AppendUint64(announcePayload, state.Downloaded)
	announcePayload = binary.BigEndian.AppendUint64(announcePayload, state.Left)
	announcePayload = binary.BigEndian.AppendUint64(announcePayload, state.Uploaded)
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, event)
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, 0)  // IP
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, 0)  // "key"
	announcePayload = binary.BigEndian.AppendUint32(announcePayload, 50) // numwant
	announcePayload = binary.BigEndian.AppendUint16(announcePayload, state.Port)

	announceResp, err := c.sendRecv(actionAnnounce, announcePayload)
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

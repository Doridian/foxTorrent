package udpproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
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

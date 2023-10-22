package udp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/Doridian/foxTorrent/pkg/tracker"
)

const (
	ActionConnect  = 0
	ActionAnnounce = 1
	ActionScrape   = 2
	ActionError    = 3
)

type UDPClient struct {
	addr string

	conn    *net.UDPConn
	udpAddr *net.UDPAddr

	connectionID uint64

	readTimeout time.Duration
	retries     int
}

func NewClient(urlParsed url.URL) (tracker.Announcer, error) {
	if urlParsed.Scheme != "udp" {
		return nil, fmt.Errorf("unsupported scheme: %s", urlParsed.Scheme)
	}

	return &UDPClient{
		addr:        urlParsed.Host,
		readTimeout: 15 * time.Second,
		retries:     3,
	}, nil
}

func (c *UDPClient) SetReadTimeout(d time.Duration) {
	c.readTimeout = d
}

func (c *UDPClient) SetRetries(r int) {
	c.retries = r
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

	for retries := 0; retries <= c.retries; retries++ {
		_, err = c.conn.WriteToUDP(packet, c.udpAddr)
		if err != nil {
			return nil, err
		}

		err = c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
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

			if recvAction == ActionError { // error
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

	connectResp, err := c.sendRecv(ActionConnect, []byte{})
	if err != nil {
		return err
	}

	if len(connectResp) < 8 {
		return errors.New("short connect response")
	}

	c.connectionID = binary.BigEndian.Uint64(connectResp[0:8])
	return nil
}

package udp_test

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/Doridian/foxTorrent/pkg/bitfield"
	"github.com/Doridian/foxTorrent/pkg/metainfo"
	"github.com/Doridian/foxTorrent/pkg/torrent/state"
	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
	"github.com/Doridian/foxTorrent/pkg/tracker/udp"
	"github.com/Doridian/foxTorrent/testfiles"
	"github.com/stretchr/testify/assert"
)

func TestAnnounceUbuntu(t *testing.T) {
	meta, err := metainfo.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoTorrent)
	assert.NoError(t, err)

	expectedAnnounce, err := announce.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounceIpv4)
	assert.NoError(t, err)

	state := &state.State{
		PeerID:     "foxTorrent dummyPeer",
		Port:       6881,
		Uploaded:   0,
		Downloaded: 0,
		Left:       meta.TotalLength(),
		InfoHash:   meta.InfoHash,
		Pieces:     bitfield.NewBitfield(uint64(len(meta.Info.Pieces))),
	}

	announceServer, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	assert.NoError(t, err)

	go func() {
		buf := make([]byte, 65536)

		var expectedConnectionID uint64
		for {
			n, recvAddr, err := announceServer.ReadFromUDP(buf)
			if err != nil {
				return
			}

			data := buf[:n]

			actualConnectionID := binary.BigEndian.Uint64(data[:8])
			action := binary.BigEndian.Uint32(data[8:12])
			transactionID := binary.BigEndian.Uint32(data[12:16])

			if action == udp.ActionConnect {
				expectedConnectionID = uint64(0x41727101980)
			}

			if actualConnectionID != expectedConnectionID {
				panic(fmt.Sprintf("Expected connection ID %x, got %x", expectedConnectionID, actualConnectionID))
			}

			switch action {
			case udp.ActionConnect:
				expectedConnectionID = uint64(time.Now().UnixNano())

				response := make([]byte, 0, 16)

				response = binary.BigEndian.AppendUint32(response, udp.ActionConnect)
				response = binary.BigEndian.AppendUint32(response, transactionID)

				response = binary.BigEndian.AppendUint64(response, expectedConnectionID)

				_, err = announceServer.WriteToUDP(response, recvAddr)
				if err != nil {
					panic(err)
				}

			case udp.ActionAnnounce:
				response := make([]byte, 0, 26)

				response = binary.BigEndian.AppendUint32(response, udp.ActionAnnounce)
				response = binary.BigEndian.AppendUint32(response, transactionID)

				response = binary.BigEndian.AppendUint32(response, expectedAnnounce.Interval)
				response = binary.BigEndian.AppendUint32(response, expectedAnnounce.Incomplete)
				response = binary.BigEndian.AppendUint32(response, expectedAnnounce.Complete)

				response = append(response, expectedAnnounce.Peers[0].IP.To4()...)
				response = binary.BigEndian.AppendUint16(response, expectedAnnounce.Peers[0].Port)

				_, err = announceServer.WriteToUDP(response, recvAddr)
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	defer announceServer.Close()

	parsedUrl := &url.URL{
		Scheme: "udp",
		Host:   announceServer.LocalAddr().String(),
		Path:   "",
	}
	client, err := udp.NewClient(*parsedUrl)
	assert.NoError(t, err)
	client.(*udp.UDPClient).SetReadTimeout(1 * time.Second)
	client.(*udp.UDPClient).SetRetries(0)

	err = client.Connect()
	assert.NoError(t, err)

	announceResp, err := client.AnnounceEvent(state, announce.EventStarted)
	assert.NoError(t, err)
	assert.NotNil(t, announceResp)

	assert.Equal(t, expectedAnnounce, announceResp)
}

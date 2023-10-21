package udpproto_test

import (
	"net"
	"testing"

	"github.com/Doridian/foxTorrent/sideband/metainfo"
	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
	"github.com/Doridian/foxTorrent/sideband/tracker/udpproto"
	"github.com/Doridian/foxTorrent/testfiles"
	"github.com/stretchr/testify/assert"
)

func TestAnnounceUbuntu(t *testing.T) {
	meta, err := metainfo.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoTorrent)
	assert.NoError(t, err)

	state := &announce.TorrentState{
		PeerID:     "foxTorrent dummyPeer",
		Port:       6881,
		Uploaded:   0,
		Downloaded: 0,
		Left:       meta.TotalLength(),
		Meta:       meta,
	}

	announceServer, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 60881,
	})
	assert.NoError(t, err)

	defer announceServer.Close()

	client, err := udpproto.NewClient("127.0.0.1:60881")
	assert.NoError(t, err)

	announceResp, err := client.AnnounceEvent(state, announce.EventStarted)
	assert.NoError(t, err)
	assert.NotNil(t, announceResp)

	expectedAnnounce, err := announce.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounceIpv4)
	assert.NoError(t, err)
	assert.Equal(t, expectedAnnounce, announceResp)
}

package announce_test

import (
	"net"
	"testing"

	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
	"github.com/Doridian/foxTorrent/testfiles"
	"github.com/stretchr/testify/assert"
)

func TestDecodeUbuntuAnnounce(t *testing.T) {
	decoded, err := announce.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounce)
	assert.NoError(t, err)

	assert.Equal(t, "", decoded.WarningMessage)

	assert.Equal(t, uint32(1800), decoded.Interval)
	assert.Equal(t, uint32(0), decoded.MinInterval)

	assert.Equal(t, "", decoded.TrackerID)

	assert.Equal(t, uint32(490), decoded.Complete)
	assert.Equal(t, uint32(13), decoded.Incomplete)

	assert.Len(t, decoded.Peers, 22)
	assert.Equal(t, net.ParseIP("2001:bc8:1864:f0f::1"), decoded.Peers[0].IP)
	assert.Equal(t, uint16(13000), decoded.Peers[0].Port)
}

func TestDecodeUbuntuAnnounceIpv4(t *testing.T) {
	decoded, err := announce.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounceIpv4)
	assert.NoError(t, err)

	assert.Equal(t, "", decoded.WarningMessage)

	assert.Equal(t, uint32(1800), decoded.Interval)
	assert.Equal(t, uint32(0), decoded.MinInterval)

	assert.Equal(t, "", decoded.TrackerID)

	assert.Equal(t, uint32(502), decoded.Complete)
	assert.Equal(t, uint32(15), decoded.Incomplete)

	assert.Len(t, decoded.Peers, 1)
	assert.Equal(t, net.ParseIP("185.125.190.59").To4(), decoded.Peers[0].IP.To4())
	assert.Equal(t, uint16(6891), decoded.Peers[0].Port)
}

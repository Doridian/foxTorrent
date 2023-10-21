package metainfo_test

import (
	"testing"
	"time"

	"github.com/Doridian/foxTorrent/sideband/metainfo"
	"github.com/Doridian/foxTorrent/testfiles"
	"github.com/stretchr/testify/assert"
)

func TestDecodeUbuntuTorrent(t *testing.T) {
	meta, err := metainfo.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoTorrent)
	assert.NoError(t, err)

	assert.Equal(t, "https://torrent.ubuntu.com/announce", meta.Announce)
	assert.Equal(t, [][]string{{"https://torrent.ubuntu.com/announce"}, {"https://ipv6.torrent.ubuntu.com/announce"}}, meta.AnnounceList)
	assert.Equal(t, "mktorrent 1.1", meta.CreatedBy)
	assert.Equal(t, time.Time(time.Date(2023, time.October, 12, 14, 24, 45, 0, time.UTC)), meta.CreationDate.UTC())
	assert.Equal(t, "Ubuntu CD releases.ubuntu.com", meta.Comment)
	assert.Equal(t, "", meta.Encoding)
	assert.Equal(t, 10156, len(meta.Info.Pieces))

	assert.Equal(t, []byte{0xc1, 0x46, 0x37, 0x92, 0xa1, 0xff, 0x36, 0xa2, 0x37, 0xe3, 0xa0, 0xf6, 0x8b, 0xad, 0xeb, 0x0d, 0x37, 0x64, 0xe9, 0xbb}, meta.InfoHash)
	assert.Equal(t, "", meta.Info.BaseName)
	assert.Equal(t, uint64(0x40000), meta.Info.PieceLength)
	assert.Equal(t, false, meta.Info.Private)
	assert.Equal(t, []metainfo.FileInfo{
		{
			Length: 2662275072,
			Path:   []string{"ubuntu-23.10-live-server-amd64.iso"},
			MD5Sum: nil,
		},
	}, meta.Info.Files)
}

package httpproto_test

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/Doridian/foxTorrent/sideband/metainfo"
	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
	"github.com/Doridian/foxTorrent/sideband/tracker/httpproto"
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

	announceServer := http.Server{
		Addr: "127.0.0.1:60881",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/announce", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			assert.Equal(t, "\xc1F7\x92\xa1\xff6\xa27\xe3\xa0\xf6\x8b\xad\xeb\r7d\xe9\xbb", r.FormValue("info_hash"))
			assert.Equal(t, "foxTorrent dummyPeer", r.FormValue("peer_id"))
			assert.Equal(t, "6881", r.FormValue("port"))
			assert.Equal(t, "1", r.FormValue("compact"))
			assert.Equal(t, "1", r.FormValue("no_peer_id"))
			assert.Equal(t, "0", r.FormValue("uploaded"))
			assert.Equal(t, "0", r.FormValue("downloaded"))
			assert.Equal(t, strconv.FormatUint(meta.TotalLength(), 10), r.FormValue("left"))
			assert.Equal(t, "started", r.FormValue("event"))
			assert.Equal(t, "50", r.FormValue("numwant"))
			assert.Equal(t, "", r.FormValue("trackerid"))

			w.Header().Set("Content-Type", "text/plain")
			w.Write(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounce)
		}),
	}
	go announceServer.ListenAndServe()
	defer announceServer.Close()

	parsedUrl, err := url.Parse("http://127.0.0.1:60881/announce")
	assert.NoError(t, err)
	client, err := httpproto.NewClient(*parsedUrl)
	assert.NoError(t, err)

	announceResp, err := client.AnnounceEvent(state, announce.EventStarted)
	assert.NoError(t, err)
	assert.NotNil(t, announceResp)

	expectedAnnounce, err := announce.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounce)
	assert.NoError(t, err)
	assert.Equal(t, expectedAnnounce, announceResp)
}

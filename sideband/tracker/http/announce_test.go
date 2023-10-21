package http_test

import (
	nethttp "net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/Doridian/foxTorrent/sideband/metainfo"
	"github.com/Doridian/foxTorrent/sideband/tracker"
	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
	"github.com/Doridian/foxTorrent/sideband/tracker/http"
	"github.com/Doridian/foxTorrent/testfiles"
	"github.com/stretchr/testify/assert"
)

func TestAnnounceUbuntu(t *testing.T) {
	meta, err := metainfo.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoTorrent)
	assert.NoError(t, err)

	state := &tracker.TorrentState{
		PeerID:     "foxTorrent dummyPeer",
		Port:       6881,
		Uploaded:   0,
		Downloaded: 0,
		Left:       meta.TotalLength(),
		Meta:       meta,
	}

	var announcRequest nethttp.Request

	announceServer := nethttp.Server{
		Addr: "127.0.0.1:60881",
		Handler: nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			if r.URL.Path != "/announce" {
				w.WriteHeader(nethttp.StatusNotFound)
				return
			}
			if r.Method != "GET" {
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
				return
			}

			announcRequest = *r
			w.Header().Set("Content-Type", "text/plain")
			w.Write(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounce)
		}),
	}
	go announceServer.ListenAndServe()
	defer announceServer.Close()

	parsedUrl, err := url.Parse("http://127.0.0.1:60881/announce")
	assert.NoError(t, err)
	client, err := http.NewClient(*parsedUrl)
	assert.NoError(t, err)

	announceResp, err := client.AnnounceEvent(state, announce.EventStarted)
	assert.NoError(t, err)
	assert.NotNil(t, announceResp)

	// Make sure the HTTP call was correct
	assert.Equal(t, "/announce", announcRequest.URL.Path)
	assert.Equal(t, "GET", announcRequest.Method)

	assert.Equal(t, "\xc1F7\x92\xa1\xff6\xa27\xe3\xa0\xf6\x8b\xad\xeb\r7d\xe9\xbb", announcRequest.FormValue("info_hash"))
	assert.Equal(t, "foxTorrent dummyPeer", announcRequest.FormValue("peer_id"))
	assert.Equal(t, "6881", announcRequest.FormValue("port"))
	assert.Equal(t, "1", announcRequest.FormValue("compact"))
	assert.Equal(t, "1", announcRequest.FormValue("no_peer_id"))
	assert.Equal(t, "0", announcRequest.FormValue("uploaded"))
	assert.Equal(t, "0", announcRequest.FormValue("downloaded"))
	assert.Equal(t, strconv.FormatUint(meta.TotalLength(), 10), announcRequest.FormValue("left"))
	assert.Equal(t, "started", announcRequest.FormValue("event"))
	assert.Equal(t, "50", announcRequest.FormValue("numwant"))
	assert.Equal(t, "", announcRequest.FormValue("trackerid"))

	// Check the response
	expectedAnnounce, err := announce.Decode(testfiles.Ubuntu2310LiveServerAMD64IsoAnnounce)
	assert.NoError(t, err)
	assert.Equal(t, expectedAnnounce, announceResp)
}

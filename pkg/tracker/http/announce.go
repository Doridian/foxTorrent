package http

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Doridian/foxTorrent/pkg/tracker"
	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
)

func (c *HTTPClient) Announce(state *tracker.TorrentState) (*announce.Announce, error) {
	return c.AnnounceEvent(state, announce.EventNone)
}

func (c *HTTPClient) AnnounceEvent(state *tracker.TorrentState, event uint32) (*announce.Announce, error) {
	useUrl := c.urlParsed

	query := useUrl.Query()
	query.Set("info_hash", string(state.Meta.InfoHash))
	query.Set("peer_id", state.PeerID)
	query.Set("port", strconv.FormatUint(uint64(state.Port), 10))
	query.Set("compact", "1")
	query.Set("no_peer_id", "1")
	query.Set("uploaded", strconv.FormatUint(state.Uploaded, 10))
	query.Set("downloaded", strconv.FormatUint(state.Downloaded, 10))
	query.Set("left", strconv.FormatUint(state.Left, 10))
	if event != announce.EventNone {
		switch event {
		case announce.EventCompleted:
			query.Set("event", "completed")
		case announce.EventStarted:
			query.Set("event", "started")
		case announce.EventStopped:
			query.Set("event", "stopped")
		}
	}
	query.Set("numwant", "50")
	query.Set("trackerid", c.trackerID)
	useUrl.RawQuery = query.Encode()

	resp, err := http.Get(useUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decoded, err := announce.Decode(data)
	if err != nil {
		return nil, err
	}

	if decoded.TrackerID != "" {
		c.trackerID = decoded.TrackerID
	}

	return decoded, err
}

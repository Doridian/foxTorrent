package httpproto

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Doridian/foxTorrent/sideband/announce"
)

type HTTPClient struct {
	urlParsed url.URL
	trackerID string
}

func NewClient(urlParsed url.URL) (announce.Announcer, error) {
	return &HTTPClient{
		urlParsed: urlParsed,
		trackerID: "",
	}, nil
}

func (c *HTTPClient) Connect() error {
	return nil
}

func (c *HTTPClient) Announce(info *announce.ClientInfo) (*announce.Announce, error) {
	return c.AnnounceEvent(info, announce.EventNone)
}

func (c *HTTPClient) AnnounceEvent(info *announce.ClientInfo, event uint32) (*announce.Announce, error) {
	useUrl := c.urlParsed

	query := useUrl.Query()
	query.Set("info_hash", string(info.Meta.InfoHash[:]))
	query.Set("peer_id", info.PeerID)
	query.Set("port", strconv.FormatUint(uint64(info.Port), 10))
	query.Set("compact", "1")
	query.Set("no_peer_id", "1")
	query.Set("uploaded", strconv.FormatUint(info.Uploaded, 10))
	query.Set("downloaded", strconv.FormatUint(info.Downloaded, 10))
	query.Set("left", strconv.FormatUint(info.Left, 10))
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

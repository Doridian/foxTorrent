package httpproto

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Doridian/foxTorrent/sideband/announce"
	"github.com/Doridian/foxTorrent/sideband/metainfo"
)

type HTTPClient struct {
	urlParsed  url.URL
	clientInfo *announce.ClientInfo
}

func NewClient(urlParsed url.URL, clientInfo *announce.ClientInfo) (announce.Announcer, error) {
	return &HTTPClient{
		urlParsed:  urlParsed,
		clientInfo: clientInfo,
	}, nil
}

func (c *HTTPClient) Connect() error {
	return nil
}

func (c *HTTPClient) Announce(meta *metainfo.Metainfo) (*announce.Announce, error) {
	return c.AnnounceEvent(meta, announce.EventNone)
}

func (c *HTTPClient) AnnounceEvent(meta *metainfo.Metainfo, event uint32) (*announce.Announce, error) {
	useUrl := c.urlParsed

	query := useUrl.Query()
	query.Set("info_hash", string(meta.InfoHash[:]))
	query.Set("peer_id", c.clientInfo.PeerID)
	query.Set("port", strconv.FormatUint(uint64(c.clientInfo.Port), 10))
	query.Set("compact", "1")
	query.Set("no_peer_id", "1")
	query.Set("uploaded", strconv.FormatUint(c.clientInfo.Uploaded, 10))
	query.Set("downloaded", strconv.FormatUint(c.clientInfo.Downloaded, 10))
	query.Set("left", strconv.FormatUint(c.clientInfo.Left, 10))
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
	query.Set("trackerid", c.clientInfo.TrackerID)
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
		c.clientInfo.TrackerID = decoded.TrackerID
	}

	return decoded, err
}

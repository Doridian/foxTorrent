package httpproto

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Doridian/foxTorrent/sideband/announce"
	"github.com/Doridian/foxTorrent/sideband/metainfo"
)

type ClientInfo struct {
	PeerID     string
	TrackerID  string
	Port       int64
	Uploaded   int64
	Downloaded int64
	Left       int64
}

func SendAnnounce(urlStr string, client *ClientInfo, meta *metainfo.Metainfo) (*announce.Announce, error) {
	return SendAnnounceEvent(urlStr, "", client, meta)
}

func SendAnnounceEvent(urlStr string, event string, client *ClientInfo, meta *metainfo.Metainfo) (*announce.Announce, error) {
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	query := urlParsed.Query()
	query.Set("info_hash", string(meta.InfoHash[:]))
	query.Set("peer_id", client.PeerID)
	query.Set("port", strconv.FormatInt(client.Port, 10))
	query.Set("compact", "1")
	query.Set("no_peer_id", "1")
	query.Set("uploaded", strconv.FormatInt(client.Uploaded, 10))
	query.Set("downloaded", strconv.FormatInt(client.Downloaded, 10))
	query.Set("left", strconv.FormatInt(client.Left, 10))
	if event != "" {
		query.Set("event", event)
	}
	query.Set("numwant", "50")
	query.Set("trackerid", client.TrackerID)
	urlParsed.RawQuery = query.Encode()

	resp, err := http.Get(urlParsed.String())
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
		client.TrackerID = decoded.TrackerID
	}

	return decoded, err
}

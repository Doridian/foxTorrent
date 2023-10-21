package httpproto

import (
	"net/url"

	"github.com/Doridian/foxTorrent/sideband/tracker/announce"
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

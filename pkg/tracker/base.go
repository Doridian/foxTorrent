package tracker

import (
	"fmt"
	"net/url"

	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
	"github.com/Doridian/foxTorrent/pkg/tracker/http"
	"github.com/Doridian/foxTorrent/pkg/tracker/udp"
)

func CreateFromURL(parsedURL url.URL) (announce.Announcer, error) {
	switch parsedURL.Scheme {
	case "http", "https":
		return http.NewClient(parsedURL)
	case "udp":
		return udp.NewClient(parsedURL)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}
}

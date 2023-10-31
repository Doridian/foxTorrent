package manager

import (
	"errors"
	"log"
	"math"
	"net/url"
	"sync"

	"github.com/Doridian/foxTorrent/pkg/bitfield"
	"github.com/Doridian/foxTorrent/pkg/metainfo"
	"github.com/Doridian/foxTorrent/pkg/torrent/state"
	"github.com/Doridian/foxTorrent/pkg/tracker"
	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
)

var ErrNoAnnouncers = errors.New("no announcers")

type TorrentManager struct {
	info    *metainfo.Metainfo
	state   *state.State
	running bool

	announcers         []announce.Announcer
	announceResults    []*announce.Announce
	announceSync       sync.WaitGroup
	announceQueueWait  sync.WaitGroup
	announceGatherLock sync.Mutex
}

func New(infoHash []byte, port uint16, peerID string) *TorrentManager {
	return &TorrentManager{
		state: &state.State{
			InfoHash:   infoHash,
			Left:       math.MaxUint64,
			Uploaded:   0,
			Downloaded: 0,
			Port:       port,
			Pieces:     bitfield.NewBitfield(0),
			PeerID:     peerID,
		},
	}
}

func (m *TorrentManager) SetInfo(info *metainfo.Metainfo) error {
	m.info = info
	m.state.Pieces = bitfield.NewBitfield(uint64(len(info.Info.Pieces)))
	m.state.Left = info.TotalLength()

	m.announcers = make([]announce.Announcer, 0)
	for _, announceUrlList := range info.AnnounceList {
		for _, announceUrl := range announceUrlList {
			parsedUrl, err := url.Parse(announceUrl)
			if err != nil {
				log.Printf("error parsing announce url: %s", err)
				continue
			}
			announcer, err := tracker.CreateFromURL(*parsedUrl)
			if err != nil {
				log.Printf("error creating announcer: %s", err)
				continue
			}
			m.announcers = append(m.announcers, announcer)
		}
	}
	m.announceResults = make([]*announce.Announce, len(m.announcers))

	if len(m.announcers) == 0 {
		return ErrNoAnnouncers
	}

	return nil
}

func (m *TorrentManager) Start() error {
	if m.running {
		return nil
	}

	m.running = true
	m.sendAnnounce(announce.EventStarted)

	m.announceSync.Add(1)
	go m.periodicAnnounce()

	return nil
}

func (m *TorrentManager) Stop() error {
	if !m.running {
		m.announceSync.Wait()
		m.announceQueueWait.Wait()
		return nil
	}

	m.running = false
	m.announceSync.Wait()

	m.sendAnnounce(announce.EventStopped)
	m.announceQueueWait.Wait()

	return nil
}

package manager

import (
	"log"
	"time"

	"github.com/Doridian/foxTorrent/pkg/tracker/announce"
)

func (m *TorrentManager) singleAnnounce(i int, event uint32) error {
	defer m.announceQueueWait.Done()

	announceResp, err := m.announcers[i].AnnounceEvent(m.state, event)
	if err != nil {
		log.Printf("error announcing to %v: %s", m.announcers[i], err)
		return err
	}
	m.announceResults[i] = announceResp

	return nil
}

func (m *TorrentManager) sendAnnounce(event uint32) error {
	m.announceGatherLock.Lock()
	defer m.announceGatherLock.Unlock()

	m.announceQueueWait.Wait()
	for i := range m.announcers {
		m.announceQueueWait.Add(1)
		go m.singleAnnounce(i, event)
	}
	return nil
}

func (m *TorrentManager) periodicAnnounce() {
	defer m.announceSync.Done()

	for m.running {
		m.announceGatherLock.Lock()
		m.announceQueueWait.Wait()
		for i := range m.announcers {
			if m.announceResults[i] != nil && m.announceResults[i].NextAnnounce.After(time.Now()) {
				continue
			}
			m.announceQueueWait.Add(1)
			go m.singleAnnounce(i, announce.EventNone)
		}
		m.announceGatherLock.Unlock()

		time.Sleep(5 * time.Second)
	}
}

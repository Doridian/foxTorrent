package metainfo

import "time"

type FileInfo struct {
	Length uint64
	MD5Sum []byte
	Path   []string
}

type InfoDict struct {
	PieceLength uint64
	Pieces      [][]byte
	Private     bool

	Files []FileInfo
}

type Metainfo struct {
	Info     InfoDict
	InfoHash []byte

	Announce     string
	AnnounceList [][]string

	CreatedBy    string
	CreationDate time.Time

	Comment string

	Encoding string
}

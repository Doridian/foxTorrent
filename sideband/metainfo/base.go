package metainfo

import (
	"errors"
)

var ErrMissingRequiredField = errors.New("missing required field")

type FileInfo struct {
	Length int64
	MD5Sum []byte
	Path   []string
}

type InfoDict struct {
	PieceLength int64
	Pieces      []byte
	Private     bool

	Name   string
	Length int64      // single-file
	MD5Sum []byte     // single-file
	Files  []FileInfo // multi-file
}

type Metainfo struct {
	Info         InfoDict
	Announce     string
	AnnounceList [][]string
	CreationDate int64
	Comment      string
	CreatedBy    string
	Encoding     string
}

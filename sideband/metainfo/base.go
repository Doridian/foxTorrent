package metainfo

type FileInfo struct {
	Length int64
	MD5Sum []byte
	Path   []string
}

type InfoDict struct {
	PieceLength int64
	Pieces      []byte
	Private     bool

	BaseName string
	Files    []FileInfo
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

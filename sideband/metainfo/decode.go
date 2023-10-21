package metainfo

import (
	"crypto/sha1"

	"github.com/Doridian/foxTorrent/sideband/bencoding"
)

func Decode(data []byte) (*Metainfo, error) {
	decoded, err := bencoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	meta := &Metainfo{}

	decodedDict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, bencoding.ErrInvalidType
	}

	announceRaw, ok := decodedDict["announce"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	announceTyped, ok := announceRaw.([]byte)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	meta.Announce = string(announceTyped)

	announceListRaw, ok := decodedDict["announce-list"]
	if ok { // optional
		announceListRawTyped, ok := announceListRaw.([]interface{})
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		announceList := make([][]string, 0, len(announceListRawTyped))
		for _, announce := range announceListRawTyped {
			announceSemiTypedList, ok := announce.([]interface{})
			if !ok {
				return nil, bencoding.ErrInvalidType
			}

			announceTypedList := make([]string, 0, len(announceSemiTypedList))
			for _, announceEntry := range announceSemiTypedList {
				announceEntryString, ok := announceEntry.([]byte)
				if !ok {
					return nil, bencoding.ErrInvalidType
				}
				announceTypedList = append(announceTypedList, string(announceEntryString))
			}
			announceList = append(announceList, announceTypedList)
		}
		meta.AnnounceList = announceList
	}

	creationDateRaw, ok := decodedDict["creation date"]
	if ok { // optional
		meta.CreationDate, ok = creationDateRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
	}

	commentRaw, ok := decodedDict["comment"]
	if ok { // optional
		commentTyped, ok := commentRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		meta.Comment = string(commentTyped)
	}

	createdByRaw, ok := decodedDict["created by"]
	if ok { // optional
		createdByTyped, ok := createdByRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		meta.CreatedBy = string(createdByTyped)
	}
	encodingRaw, ok := decodedDict["encoding"]
	if ok { // optional
		encodingTyped, ok := encodingRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		meta.Encoding = string(encodingTyped)
	}

	infoDictRaw, ok := decodedDict["info"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}

	infoDictTyped := InfoDict{}

	infoDict, ok := infoDictRaw.(map[string]interface{})
	if !ok {
		return nil, bencoding.ErrInvalidType
	}

	infoDictMeta := infoDict[bencoding.DictMetaEntry].(bencoding.DictMeta)
	meta.InfoHash = sha1.Sum(data[infoDictMeta.Begin:infoDictMeta.End])

	pieceLengthRaw, ok := infoDict["piece length"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	infoDictTyped.PieceLength, ok = pieceLengthRaw.(int64)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	piecesRaw, ok := infoDict["pieces"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	infoDictTyped.Pieces, ok = piecesRaw.([]byte)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	privateRaw, ok := infoDict["private"]
	if ok { // optional
		privateTyped, ok := privateRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		infoDictTyped.Private = privateTyped == 1
	}

	nameRaw, ok := infoDict["name"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	nameTyped, ok := nameRaw.([]byte)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	infoDictTyped.BaseName = string(nameTyped)

	lengthRaw, ok := infoDict["length"]
	if ok { // optional, indicates single-file mode
		singleFile := FileInfo{
			Path: []string{},
		}
		singleFile.Length, ok = lengthRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		md5sumRaw, ok := infoDict["md5sum"]
		if ok { // optional
			singleFile.MD5Sum, ok = md5sumRaw.([]byte)
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
		}
		infoDictTyped.Files = []FileInfo{singleFile}
	} else { // optional, but if missing must mean multi-file mode!
		filesRaw, ok := infoDict["files"]
		if !ok { // required
			return nil, bencoding.ErrMissingRequiredField
		}
		filesRawTyped, ok := filesRaw.([]interface{})
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		files := make([]FileInfo, 0, len(filesRawTyped))
		for _, fileRaw := range filesRawTyped {
			fileRawTyped, ok := fileRaw.(map[string]interface{})
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
			file := FileInfo{}

			lengthRaw, ok := fileRawTyped["length"]
			if !ok { // required
				return nil, bencoding.ErrMissingRequiredField
			}
			file.Length, ok = lengthRaw.(int64)
			if !ok {
				return nil, bencoding.ErrInvalidType
			}

			md5sumRaw, ok := fileRawTyped["md5sum"]
			if ok { // optional
				file.MD5Sum, ok = md5sumRaw.([]byte)
				if !ok {
					return nil, bencoding.ErrInvalidType
				}
			}

			pathRaw, ok := fileRawTyped["path"]
			if !ok { // required
				return nil, bencoding.ErrMissingRequiredField
			}
			pathRawTyped, ok := pathRaw.([]interface{})
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
			path := make([]string, 0, len(pathRawTyped))
			for _, pathEntry := range pathRawTyped {
				pathEntryString, ok := pathEntry.([]byte)
				if !ok {
					return nil, bencoding.ErrInvalidType
				}
				path = append(path, string(pathEntryString))
			}
			file.Path = path

			files = append(files, file)
		}
		meta.Info.Files = files
	}

	meta.Info = infoDictTyped

	return meta, nil
}

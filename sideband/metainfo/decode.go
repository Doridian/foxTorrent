package metainfo

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/Doridian/foxTorrent/bencoding"
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
		creationDateTyped, ok := creationDateRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		meta.CreationDate = time.Unix(creationDateTyped, 0)
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
	sha1Sum := sha1.Sum(data[infoDictMeta.Begin:infoDictMeta.End])
	meta.InfoHash = sha1Sum[:]

	pieceLengthRaw, ok := infoDict["piece length"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	pieceLengthTyped, ok := pieceLengthRaw.(int64)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	infoDictTyped.PieceLength = uint64(pieceLengthTyped)

	piecesRaw, ok := infoDict["pieces"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	piecesTyped, ok := piecesRaw.([]byte)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	if len(piecesTyped)%20 != 0 {
		return nil, fmt.Errorf("invalid pieces length: %d", len(piecesTyped))
	}

	infoDictTyped.Pieces = make([][]byte, 0, len(piecesTyped)/20)
	for i := 0; i < len(piecesTyped); i += 20 {
		infoDictTyped.Pieces = append(infoDictTyped.Pieces, piecesTyped[i:i+20])
	}

	privateRaw, ok := infoDict["private"]
	if ok { // optional
		privateTyped, ok := privateRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		infoDictTyped.Private = privateTyped == 1
	}

	baseNameRaw, ok := infoDict["name"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	baseNameTyped, ok := baseNameRaw.([]byte)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}

	lengthRaw, ok := infoDict["length"]
	if ok { // optional, indicates single-file mode
		singleFile := FileInfo{
			Path: []string{string(baseNameTyped)},
		}
		lengthTyped, ok := lengthRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		singleFile.Length = uint64(lengthTyped)

		md5sumRaw, ok := infoDict["md5sum"]
		if ok { // optional
			singleFile.MD5Sum, ok = md5sumRaw.([]byte)
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
		}

		infoDictTyped.Files = []FileInfo{singleFile}
	} else { // optional, but if missing must mean multi-file mode!
		baseNameString := string(baseNameTyped)

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
			lengthTyped, ok := lengthRaw.(int64)
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
			file.Length = uint64(lengthTyped)

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
			path := make([]string, 0, len(pathRawTyped)+1)
			path = append(path, baseNameString)
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
		infoDictTyped.Files = files
	}

	meta.Info = infoDictTyped

	return meta, nil
}

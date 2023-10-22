package bencoding

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var ErrEndOfItem = errors.New("end-of-item")

const DictMetaEntry = "$$meta$$"

type DictMeta struct {
	Begin int64
	End   int64
}

func getDecoderPosition(reader *bytes.Reader) int64 {
	pos, _ := reader.Seek(0, io.SeekCurrent)
	return pos
}

func makeDecoderError(reader *bytes.Reader, err string) error {
	return fmt.Errorf("%s at %d", err, getDecoderPosition(reader))
}

func readNumeric(reader *bytes.Reader, terminator byte) (int64, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return readNumericWithFirstChar(reader, b, terminator)
}

func readNumericWithFirstChar(reader *bytes.Reader, first byte, terminator byte) (int64, error) {
	var res int64
	var multiplier int64 = 1

	if first == '-' {
		multiplier = -1
	} else if first == '0' {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		if b != terminator {
			return 0, makeDecoderError(reader, fmt.Sprintf("encountered unexpected \"%c\" after leading zero while trying to read numeric", b))
		}
		return 0, nil
	} else if first < '1' || first > '9' {
		return 0, makeDecoderError(reader, fmt.Sprintf("encountered unexpected \"%c\" while trying to read first byte of numeric", first))
	} else {
		res = int64(first - '0')
	}

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		if b == '0' && res == 0 {
			return 0, makeDecoderError(reader, "encountered leading zero while reading numeric")
		}

		if b < '0' || b > '9' {
			if b != terminator {
				return 0, makeDecoderError(reader, fmt.Sprintf("encountered unexpected \"%c\" while trying to read numeric", b))
			}
			if res == 0 {
				return 0, makeDecoderError(reader, "encountered numeric terminator without reading any digits")
			}
			return res * multiplier, nil
		}

		res = res*10 + int64(b-'0')
	}
}

func decodeReader(reader *bytes.Reader) (interface{}, error) {
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, nil
			}
			return nil, err
		}

		if b == 'i' { // integer
			return readNumeric(reader, 'e')
		} else if b == 'l' { // list
			res := make([]interface{}, 0)
			for {
				item, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if item == nil {
					return nil, makeDecoderError(reader, "encountered EOF trying to decode item")
				}
				if item == ErrEndOfItem {
					return res, nil
				}
				res = append(res, item)
			}
		} else if b == 'd' { // dict string => any
			res := make(map[string]interface{})
			begin := getDecoderPosition(reader) - 1
			for {
				key, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if key == nil {
					return nil, makeDecoderError(reader, "encountered EOF trying to decode key")
				}
				if key == ErrEndOfItem {
					res[DictMetaEntry] = DictMeta{
						Begin: begin,
						End:   getDecoderPosition(reader),
					}
					return res, nil
				}
				keyStr, ok := key.([]byte)
				if !ok {
					return nil, makeDecoderError(reader, fmt.Sprintf("dict key \"%v\" is not a string", key))
				}
				value, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if value == nil {
					return nil, makeDecoderError(reader, "encountered EOF trying to decode value")
				}
				if value == ErrEndOfItem {
					return nil, makeDecoderError(reader, "encountered end-of-item trying to decode value")
				}
				res[string(keyStr)] = value
			}
		} else if b >= '0' && b <= '9' { // string
			err = reader.UnreadByte()
			if err != nil {
				return nil, err
			}
			length, err := readNumeric(reader, ':')
			if err != nil {
				return nil, err
			}
			if length < 0 {
				return nil, makeDecoderError(reader, fmt.Sprintf("encountered negative length while reading string (got %d)", length))
			}

			if length == 0 {
				return []byte{}, nil
			}

			buf := make([]byte, length)
			readLength, err := reader.Read(buf)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}
			if readLength != int(length) || errors.Is(err, io.EOF) {
				return nil, makeDecoderError(reader, fmt.Sprintf("encountered short read while reading string (got %d, expected %d)", readLength, length))
			}

			return buf, nil
		} else if b == 'e' { // end-of-item
			return ErrEndOfItem, nil
		} else {
			return nil, makeDecoderError(reader, fmt.Sprintf("encountered unexpected \"%c\" while trying to read next element", b))
		}
	}
}

func DecodeString(input []byte) (interface{}, error) {
	reader := bytes.NewReader(input)
	obj, err := decodeReader(reader)
	if obj == ErrEndOfItem {
		return nil, errors.New("unexpected end-of-item")
	}
	return obj, err
}

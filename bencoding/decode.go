package bencoding

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var ErrEndOfItem = errors.New("end-of-item")

func makeDecoderError(reader *bytes.Reader, err string) error {
	pos, _ := reader.Seek(0, io.SeekCurrent)
	return fmt.Errorf("%s at %d", err, pos)
}

func readNumeric(reader *bytes.Reader, terminator byte, allowNegative bool) (int64, error) {
	var res int64 = 0
	var multiplier int64 = 1

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		if b == '-' && multiplier == 1 && res == 0 {
			multiplier = -1
			continue
		}

		if b == '0' {
			if multiplier != 1 {
				return 0, makeDecoderError(reader, "encountered leading zero in numeric")
			}

			b, err = reader.ReadByte()
			if err != nil {
				return 0, err
			}
			if b != terminator {
				return 0, makeDecoderError(reader, fmt.Sprintf("encountered unexpected \"%c\" while trying to read numeric", b))
			}
			return 0, nil
		}

		if b < '1' || b > '9' {
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
			return readNumeric(reader, 'e', true)
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
			for {
				key, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if key == nil {
					return nil, makeDecoderError(reader, "encountered EOF trying to decode key")
				}
				if key == ErrEndOfItem {
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
			length, err := readNumeric(reader, ':', false)
			if err != nil {
				return nil, err
			}
			buf := make([]byte, length)
			if length > 0 {
				readLength, err := reader.Read(buf)
				if err != nil && !errors.Is(err, io.EOF) {
					return nil, err
				}
				if readLength != int(length) || errors.Is(err, io.EOF) {
					return nil, makeDecoderError(reader, fmt.Sprintf("encountered short read while reading string (got %d, expected %d)", readLength, length))
				}
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
